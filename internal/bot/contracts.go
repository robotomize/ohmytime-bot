package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	AnswerCallbackQuery(config tgbotapi.CallbackConfig) (tgbotapi.APIResponse, error)
}
