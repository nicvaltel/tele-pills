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

// var bbb = tgbotapi.NewInlineKeyboardButtonData("Добавить напоминание", "add_reminder")

// msg := tgbotapi.NewMessage(chatId, "Добавить напоминание:")
// 	button := tgbotapi.NewInlineKeyboardButtonData("Добавить напоминание", "add_reminder")
// 	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
// 		tgbotapi.NewInlineKeyboardRow(button),
// 	)
// 	msg.ReplyMarkup = inlineKeyboard
