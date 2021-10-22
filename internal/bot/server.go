package bot

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/enescakir/emoji"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/robotomize/ohmytimebot/internal/logging"
)

const SendingMessageError = "Oops, something went wrong"

type Options struct {
	PollingTimeout    int
	UpdatesMaxWorkers int
}

type Option func(*Dispatcher)

func NewDispatcher(cfg Config, opts ...Option) (*Dispatcher, error) {
	idx, err := bleve.Open(cfg.PathToIndex)
	if err != nil {
		return nil, fmt.Errorf("bleve open: %w", err)
	}

	d := Dispatcher{
		opts: Options{
			PollingTimeout:    cfg.PollingTimeout,
			UpdatesMaxWorkers: cfg.MaxWorkers,
		},
		indexer: idx,
	}

	for _, o := range opts {
		o(&d)
	}

	return &d, nil
}

type Dispatcher struct {
	opts    Options
	indexer bleve.Index
}

func (s *Dispatcher) Run(ctx context.Context, telegram *tgbotapi.BotAPI, cfg Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	updates, err := s.setupTelegramMode(ctx, telegram, cfg)
	if err != nil {
		return fmt.Errorf("configuring telegram updates: %w", err)
	}

	go func() {
		<-ctx.Done()
		telegram.StopReceivingUpdates()
	}()

	var wg sync.WaitGroup

	for i := 0; i < s.opts.UpdatesMaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.dispatchingMessages(ctx, telegram, updates)
		}()
	}

	wg.Wait()

	return nil
}

func (s *Dispatcher) setupTelegramMode(ctx context.Context, telegram *tgbotapi.BotAPI, cfg Config) (tgbotapi.UpdatesChannel, error) {
	logger := logging.FromContext(ctx).Named("Dispatcher.setupTelegramMode")
	if cfg.WebHookURL != "" {
		if _, err := telegram.SetWebhook(tgbotapi.NewWebhook(cfg.WebHookURL + cfg.Token)); err != nil {
			return nil, fmt.Errorf("telegram set webhook: %w", err)
		}
		info, err := telegram.GetWebhookInfo()
		if err != nil {
			return nil, fmt.Errorf("telegram get webhook info: %w", err)
		}

		if info.LastErrorDate != 0 {
			logger.Errorf("Telegram callback failed: %s", info.LastErrorMessage)
		}

		updates := telegram.ListenForWebhook("/" + cfg.Token)
		go func() {
			if err = http.ListenAndServe(cfg.WebHookURL, nil); err != nil {
				logger.Fatalf("Listen and serve http stopped: %v", err)
			}
		}()

		return updates, nil
	}

	resp, err := telegram.RemoveWebhook()
	if err != nil {
		return nil, fmt.Errorf("telegram client remove webhook: %w", err)
	}

	if !resp.Ok {
		if resp.ErrorCode > 0 {
			return nil, fmt.Errorf(
				"telegram client remove webhook with error code %d and description %s",
				resp.ErrorCode, resp.Description,
			)
		}

		return nil, fmt.Errorf("telegram client remove webhook response not ok")
	}

	updatesChanConfig := tgbotapi.NewUpdate(0)
	updatesChanConfig.Timeout = s.opts.PollingTimeout
	updates, err := telegram.GetUpdatesChan(updatesChanConfig)
	if err != nil {
		return nil, fmt.Errorf("telegram get updates chan: %w", err)
	}

	return updates, nil
}

func (s *Dispatcher) handleMessage(ctx context.Context, sender TelegramSender, message *tgbotapi.Message) error {
	logger := logging.FromContext(ctx).Named("Dispatcher.handleMessage")
	query := bleve.NewQueryStringQuery(message.Text)
	searchRequest := bleve.NewSearchRequest(query)
	result, err := s.indexer.Search(searchRequest)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	if result.Hits.Len() == 0 {
		if _, err := sender.Send(
			tgbotapi.NewMessage(
				message.Chat.ID, fmt.Sprintf("%s Unfortunately, nothing was found", emoji.Robot.String()),
			),
		); err != nil {
			return fmt.Errorf("send msg: %w", err)
		}

		return nil
	}

	if result.Hits.Len() > 1 {
		config := tgbotapi.NewMessage(
			message.Chat.ID, fmt.Sprintf(
				"%s I got something! Pick a location! %s", emoji.Robot.String(), emoji.DirectHit.String(),
			),
		)
		markup := tgbotapi.NewInlineKeyboardMarkup()

		row := tgbotapi.NewInlineKeyboardRow()
		for _, hit := range result.Hits {
			if len(row) == 3 {
				markup.InlineKeyboard = append(markup.InlineKeyboard, row)
				row = tgbotapi.NewInlineKeyboardRow()
			}

			doc, err := s.indexer.Document(hit.ID)
			if err != nil {
				logger.Errorf("can not open document by ID %s: %v", hit.ID, err)
			}
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s(%s)", doc.Fields[0].Value(), doc.Fields[3].Value()), hit.ID),
			)
		}

		if len(row) > 0 {
			markup.InlineKeyboard = append(markup.InlineKeyboard, row)
		}

		config.ReplyMarkup = markup

		if _, err = sender.Send(config); err != nil {
			return fmt.Errorf("send msg: %w", err)
		}

		return nil
	}

	doc, err := s.indexer.Document(result.Hits[0].ID)
	if err != nil {
		return fmt.Errorf("document: %w", err)
	}
	tz := doc.Fields[2]
	if err = s.sendTime(string(tz.Value()), fmt.Sprintf("%s(%s)", doc.Fields[0].Value(), doc.Fields[3].Value()), message.Chat.ID, sender); err != nil {
		return fmt.Errorf("send time: %w", err)
	}

	return nil
}

const StartCommandText = "start"

var StartCommandMessage = "Hi, this is a bot" + emoji.Robot.String() + " that shows the local time of the selected location\n" +
	"Write the name of the location and get result" +
	"\n\n*source code:* [github](https://github.com/robotomize/ohmytime-bot)"

func (s *Dispatcher) dispatchingMessages(ctx context.Context, sender TelegramSender, updates tgbotapi.UpdatesChannel) {
	logger := logging.FromContext(ctx).Named("Dispatcher.dispatchingMessages")
	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				cmd := update.Message.Command()
				if cmd == StartCommandText {
					config := tgbotapi.NewMessage(update.Message.Chat.ID, StartCommandMessage)
					config.ParseMode = tgbotapi.ModeMarkdown
					if _, err := sender.Send(config); err != nil {
						logger.Errorf("send message: %v", err)
					}
				}
				continue
			}

			if err := s.handleMessage(ctx, sender, update.Message); err != nil {
				logger.Errorf("handle telegram message: %v", err)
			}
		}

		if update.CallbackQuery != nil {
			idx := update.CallbackQuery.Data
			if _, err := sender.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, idx)); err != nil {
				logger.Errorf("answer callback: %v", err)
				continue
			}

			doc, err := s.indexer.Document(idx)
			if err != nil {
				logger.Errorf("indexer document: %v", err)
				if _, err = sender.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, SendingMessageError)); err != nil {
					logger.Errorf("send msg: %v", err)
				}
			}

			tz := doc.Fields[2]
			if err = s.sendTime(
				string(tz.Value()),
				fmt.Sprintf("%s(%s)", doc.Fields[0].Value(), doc.Fields[3].Value()),
				update.CallbackQuery.Message.Chat.ID, sender,
			); err != nil {
				logger.Errorf("send time: %v", err)
			}
		}
	}
}

func (s *Dispatcher) sendTime(tz string, locationID string, chatID int64, sender TelegramSender) error {
	location, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("load location %s: %w", tz, err)
	}

	localtime := time.Now().In(location)
	msg := fmt.Sprintf(
		"Hey %s, local time in your location *%s*:\n\n%s *Date:* %s\n\n%s *Time:* %s", emoji.Robot.String(),
		locationID, emoji.Calendar.String(), localtime.Format("02 Jan 2006"), emoji.Stopwatch.String(),
		localtime.Format("15:04"),
	)
	config := tgbotapi.NewMessage(chatID, msg)
	config.ParseMode = tgbotapi.ModeMarkdown

	if _, err = sender.Send(config); err != nil {
		return fmt.Errorf("send msg: %w", err)
	}

	return nil
}
