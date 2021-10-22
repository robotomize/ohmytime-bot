package main

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/robotomize/ohmytimebot/buildinfo"
	"github.com/robotomize/ohmytimebot/internal/bot"
	"github.com/robotomize/ohmytimebot/internal/logging"
	"github.com/robotomize/ohmytimebot/internal/shutdown"
	"github.com/sethvargo/go-envconfig"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Addr     string `env:"ADDR,default=localhost:8080"`
	LogLevel string `env:"LOG_LEVEL,default=debug"`
	Telegram bot.Config
}

func main() {
	fmt.Fprintf(os.Stdout, buildinfo.Graffiti)
	_, _ = fmt.Fprintf(
		os.Stdout,
		buildinfo.GreetingCLI,
		buildinfo.Info.Tag(),
		buildinfo.Info.Time(),
		buildinfo.TgBloopURL,
		buildinfo.GithubBloopURL,
	)

	ctx, cancel := shutdown.New()
	defer cancel()

	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatalf("envconfig.Process: %v", err)
	}

	logger := logging.NewLogger(cfg.LogLevel).
		With("build_tag", buildinfo.Info.Tag()).
		With("build_time", buildinfo.Info.Time())
	ctx = logging.WithLogger(ctx, logger)

	mux := http.NewServeMux()
	mux.HandleFunc(
		"/health", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"status": "ok"}`)
		},
	)
	mux.Handle("/debug/pprof/", http.Handler(http.DefaultServeMux))

	go func() {
		if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
			logger.Errorf("listen and serve metrics: %v", err)
			cancel()
		}
	}()

	tg, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		logger.Fatalf("new telegram bot api: %v", err)
	}

	dispatcher, err := bot.NewDispatcher(cfg.Telegram)
	if err != nil {
		if errors.Is(err, bleve.ErrorIndexPathDoesNotExist) {
			logger.Fatalf("index file not found, set variable PATH_TO_INDEX, index folder is in the ./internal/search/cities.idx")
		}

		logger.Fatalf("new bot dispatcher: %v", err)
	}

	if err = dispatcher.Run(ctx, tg, cfg.Telegram); err != nil {
		logger.Fatalf("dispatcher run: %v", err)
	}
}
