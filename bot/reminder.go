package bot

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (tb TeleBot) reminderRoutine() error {
	rems, err := tb.repo.GetReminds(time.Now())
	if err != nil {
		return err
	}

	for _, rem := range rems {
		unread, remExist := tb.unreadReminders[rem.ChatId]
		if !remExist {
			tb.unreadReminders[rem.ChatId] = make(map[int64]time.Time)

			msg := tgbotapi.NewMessage(rem.ChatId, rem.ReminderMsg)
			button := tgbotapi.NewInlineKeyboardButtonData("Сделано", fmt.Sprintf("reminder_done_%d", rem.ReminderId))
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(button),
			)
			msg.ReplyMarkup = inlineKeyboard
			tb.bot.Send(msg)
			(tb.unreadReminders[rem.ChatId])[rem.ReminderId] = time.Now()
		} else if unread[rem.ReminderId].Add(tb.delayBetweenReminds).Compare(time.Now()) < 0 {
			msg := tgbotapi.NewMessage(rem.ChatId, rem.ReminderMsg)
			button := tgbotapi.NewInlineKeyboardButtonData("Сделано", fmt.Sprintf("reminder_done_%d", rem.ReminderId))
			inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(button),
			)
			msg.ReplyMarkup = inlineKeyboard
			tb.bot.Send(msg)
			(tb.unreadReminders[rem.ChatId])[rem.ReminderId] = time.Now()
		}
	}
	return nil
}
