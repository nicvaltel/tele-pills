package bot

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (tb TeleBot) handleMessageNormal(update tgbotapi.Update) error {
	return tb.handleDefault(update.Message.Chat.ID)
}

func (tb TeleBot) handleDefault(chatId ChatID) error {
	msg := tgbotapi.NewMessage(chatId, "Добавить напоминание:")
	button := tgbotapi.NewInlineKeyboardButtonData("Добавить напоминание", "add_reminder")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)
	msg.ReplyMarkup = inlineKeyboard
	tb.bot.Send(msg)
	return nil
}

func (tb TeleBot) handleCallbackQueryNormal(callback *tgbotapi.CallbackQuery) error {
	if callback.Data == "add_reminder" {
		msg := tgbotapi.NewMessage(callback.Message.Chat.ID, "Напишите название препарата:")
		tb.bot.Send(msg)
		tb.chatStates[callback.Message.Chat.ID] = BotStateWaitForInputPillName
		return nil
	} else {
		return fmt.Errorf("handleCallbackQuery can't parse CallbackData = %v", callback.Data)
	}
}

func (tb TeleBot) handleMessageInputPillName(update tgbotapi.Update) error {

	sendSetHourButton := func(chatId ChatID) {
		msg := tgbotapi.NewMessage(chatId, "Выберите час:")
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup()
		var rows [][]tgbotapi.InlineKeyboardButton

		for i := 0; i < 6; i++ {
			var row []tgbotapi.InlineKeyboardButton
			for j := 0; j < 4; j++ {
				number := i*4 + j
				callbackData := fmt.Sprintf("hour_%02d", number)
				button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d", number), callbackData)
				row = append(row, button)
			}
			rows = append(rows, row)
		}

		inlineKeyboard.InlineKeyboard = rows
		msg.ReplyMarkup = inlineKeyboard
		tb.bot.Send(msg)
		tb.chatStates[chatId] = BotStateWaitForInputHour
	}

	chatId := update.Message.Chat.ID

	if update.Message.Text == "" {
		return fmt.Errorf("Empty pill Name")
	}

	tb.pillData[chatId] = PillData{
		pillName: update.Message.Text,
		hour:     0,
		min:      0,
	}

	sendSetHourButton(chatId)
	return nil
}

func (tb TeleBot) handleCallbackQueryInputPillName(callback *tgbotapi.CallbackQuery) error {
	tb.stateToNormal(callback.Message.Chat.ID)
	return tb.handleCallbackQueryNormal(callback) // it's OK - no callback query expected, if so handle with normal handle
}

func (tb TeleBot) handleMessageInputHour(update tgbotapi.Update) error {
	tb.stateToNormal(update.Message.Chat.ID)
	return tb.handleDefault(update.Message.Chat.ID) // no message expected, if so process via default handler
}

func (tb TeleBot) handleCallbackQueryInputHour(callback *tgbotapi.CallbackQuery) error {
	sendSetMinuteButton := func(chatId ChatID) {
		msg := tgbotapi.NewMessage(chatId, "Выберите минуту:")

		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup()

		var rows [][]tgbotapi.InlineKeyboardButton
		for i := 0; i < 3; i++ {
			var row []tgbotapi.InlineKeyboardButton
			for j := 0; j < 4; j++ {
				number := (i*4 + j) * 5
				callbackData := fmt.Sprintf("minute_%02d", number)
				button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d", number), callbackData)
				row = append(row, button)
			}
			rows = append(rows, row)
		}
		inlineKeyboard.InlineKeyboard = rows

		msg.ReplyMarkup = inlineKeyboard

		tb.bot.Send(msg)

		tb.chatStates[chatId] = BotStateWaitForInputMinute

	}

	calbackStr := callback.Data
	chatId := callback.Message.Chat.ID

	if strings.HasPrefix(calbackStr, "hour_") {
		lastTwo := calbackStr[len(calbackStr)-2:]
		hour, err := strconv.Atoi(lastTwo)
		if err != nil {
			return fmt.Errorf("handleCallbackQuery can't parse hour: %s", calbackStr)
		}
		tData := tb.pillData[chatId]
		tData.hour = uint8(hour)
		tb.pillData[chatId] = tData
		sendSetMinuteButton(chatId)
		return nil
	} else {
		return tb.handleDefault(chatId)
	}
}

func (tb TeleBot) handleMessageInputMinute(update tgbotapi.Update) error {
	tb.stateToNormal(update.Message.Chat.ID)
	return tb.handleDefault(update.Message.Chat.ID) // no message expected, if so process via default handler
}

func (tb TeleBot) handleCallbackQueryInputMinute(callback *tgbotapi.CallbackQuery) error {
	addReminderToRepo := func(chatId ChatID, pillName string, hour uint8, min uint8) error {
		err := tb.repo.SaveReminder(chatId, pillName, hour, min)
		if err != nil {
			return err
		}
		msgString := fmt.Sprintf("Добавлено напоминание: %s в %02d:%02d", pillName, hour, min)
		msg := tgbotapi.NewMessage(chatId, msgString)
		tb.bot.Send(msg)
		tb.stateToNormal(chatId)
		tb.handleDefault(chatId)
		return nil
	}

	calbackStr := callback.Data
	chatId := callback.Message.Chat.ID

	if strings.HasPrefix(calbackStr, "minute_") {
		lastTwo := calbackStr[len(calbackStr)-2:]
		min, err := strconv.Atoi(lastTwo)
		if err != nil {
			return fmt.Errorf("handleCallbackQuery can't parse minute: %s", calbackStr)
		}
		tData := tb.pillData[chatId]
		tb.stateToNormal(chatId)
		return addReminderToRepo(chatId, tData.pillName, tData.hour, uint8(min))
	} else {
		tb.stateToNormal(chatId)
		return tb.handleDefault(chatId)
	}
}

func (tb TeleBot) handleAllMessages(update tgbotapi.Update, handlerFunc func(update tgbotapi.Update) error) error {

	sendWelcomeMessage := func(chatId ChatID) {
		msgString := fmt.Sprintf("Я бот-напоминатель о приёме лекарств. Чтобы добавить напоминание отправьте любое сообщение.")
		msg := tgbotapi.NewMessage(chatId, msgString)
		tb.bot.Send(msg)
	}

	isNewcomer, err := tb.repo.UserIsNewcomer(update.Message.Chat.ID)
	if err != nil {
		return err
	}

	if isNewcomer {
		sendWelcomeMessage(update.Message.Chat.ID)
		err := tb.repo.SaveUser(
			update.Message.Chat.ID,
			update.Message.Chat.UserName,
			update.Message.Chat.FirstName,
			update.Message.Chat.LastName,
			update.Message.Time(),
		)
		if err != nil {
			return err
		}
	}

	err = tb.repo.SaveMessage(
		update.Message.Chat.ID,
		update.Message.Text,
		false,
		update.Message.Time(),
	)
	if err != nil {
		return err
	}

	return handlerFunc(update)
}

func (tb TeleBot) handleAllhandleCallbackQuerys(callback *tgbotapi.CallbackQuery, handlerFunc func(callback *tgbotapi.CallbackQuery) error) error {

	err := tb.repo.SaveMessage(
		callback.Message.Chat.ID,
		callback.Data,
		true,
		callback.Message.Time(),
	)
	if err != nil {
		return err
	}

	remIdStr, foundReminderDone := strings.CutPrefix(callback.Data, "reminder_done_")
	if foundReminderDone {
		return tb.handleReminderDone(remIdStr, callback.Message.Chat.ID)
	} else {
		return handlerFunc(callback)
	}
}

func (tb TeleBot) handleReminderDone(remIdStr string, chatId ChatID) error { //TODO remind once in 5 minutes, not every 1 minute
	remId, err := strconv.Atoi(remIdStr)
	if err != nil {
		return fmt.Errorf("Incorrect reminder_done_ callback: %d", chatId)
	} else {
		if _, isPresent := (tb.unreadReminders[chatId])[int64(remId)]; isPresent {
			delete(tb.unreadReminders[chatId], int64(remId))
			err := tb.repo.UpdateRemind(int64(remId))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
