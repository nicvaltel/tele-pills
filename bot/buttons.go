package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

var menuKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Добавить лекарство"),
		tgbotapi.NewKeyboardButton("Удалить лекарство"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("XXX"),
		tgbotapi.NewKeyboardButton("YYY"),
	),
)
