package bot

import (
	"Pills/database/postgresql"
	"Pills/mdl"
	"Pills/utls"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ChatID = int64
type BotState uint8

const (
	BotStateNormal               BotState = 0
	BotStateWaitForInputPillName BotState = 1
	BotStateWaitForInputHour     BotState = 2
	BotStateWaitForInputMinute   BotState = 3
)

type TeleBot struct {
	bot        *tgbotapi.BotAPI
	repo       mdl.Repo
	chatStates map[ChatID]BotState
	tmpData    map[ChatID]TmpData
}

type TmpData struct {
	pillName string
	hour     uint8
	min      uint8
}

func (tb TeleBot) handleMessage(update tgbotapi.Update) error {

	isNewcomer, err := tb.repo.UserIsNewcomer(update.Message.Chat.ID)
	if err != nil {
		return err
	}

	if isNewcomer {
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
		update.Message.Time(),
	)
	if err != nil {
		return err
	}

	// tb.chatStates[update.Message.Chat.ID] = BotStateNormal

	switch tb.chatStates[update.Message.Chat.ID] {
	case BotStateNormal:
		err = tb.handleMessageNormal(update)
	case BotStateWaitForInputPillName:
		err = tb.handleMessageParsePillName(update)
	case BotStateWaitForInputHour:
	// err = tb.handle
	default:
		log.Printf("handleMessage incorrect BotState: %v", tb.chatStates[update.Message.Chat.ID])
	}

	return err
}

func (tb TeleBot) handleMessageNormal(update tgbotapi.Update) error {
	tb.sendAddReminderButton(update.Message.Chat.ID)
	return nil
}

func (tb TeleBot) handleMessageParsePillName(update tgbotapi.Update) error {

	if update.Message.Text == "" {
		tb.chatStates[update.Message.Chat.ID] = BotStateNormal
		return fmt.Errorf("Empty pill Name")
	}

	tb.tmpData[update.Message.Chat.ID] = TmpData{
		pillName: update.Message.Text,
		hour:     0,
		min:      0,
	}

	tb.sendSetHourButton(update.Message.Chat.ID)
	return nil
}

func (tb TeleBot) sendAddReminderButton(chatId ChatID) {
	msg := tgbotapi.NewMessage(chatId, "Добавить напоминание:")
	button := tgbotapi.NewInlineKeyboardButtonData("Добавить напоминание", "add_reminder")
	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)

	msg.ReplyMarkup = inlineKeyboard

	tb.bot.Send(msg)
}

func (tb TeleBot) sendSetMinuteButton(chatId ChatID) {
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

func (tb TeleBot) sendSetHourButton(chatId ChatID) {
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

func (tb TeleBot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) error {
	calbackStr := callback.Data
	chatId := callback.Message.Chat.ID

	if strings.HasPrefix(calbackStr, "hour_") {
		lastTwo := calbackStr[len(calbackStr)-2:]
		hour, err := strconv.Atoi(lastTwo)
		if err != nil {
			return fmt.Errorf("handleCallbackQuery can't parse hour: %s", calbackStr)
		}
		tData := tb.tmpData[chatId]
		tData.hour = uint8(hour)
		tb.tmpData[chatId] = tData
		tb.sendSetMinuteButton(chatId)
		return nil
	}

	if strings.HasPrefix(calbackStr, "minute_") {
		lastTwo := calbackStr[len(calbackStr)-2:]
		min, err := strconv.Atoi(lastTwo)
		if err != nil {
			return fmt.Errorf("handleCallbackQuery can't parse minute: %s", calbackStr)
		}
		tData := tb.tmpData[chatId]
		// tData.min = uint8(min)
		// tb.tmpData[chatId] = tData
		tb.addReminderToRepo(chatId, tData.pillName, tData.hour, uint8(min))
		delete(tb.tmpData, chatId)

		return nil
	}

	switch calbackStr {
	case "add_reminder":
		tb.addReminder(chatId)
	default:
		return fmt.Errorf("handleCallbackQuery can't parse CallbackData = %v", callback.Data)
	}
	return nil
}

func (tb TeleBot) addReminderToRepo(chatId ChatID, pillName string, hour uint8, min uint8) error {
	err := tb.repo.SaveReminder(chatId, pillName, hour, min)
	if err != nil {
		return err
	}
	msgString := fmt.Sprintf("Добавлено напоминание: %s в %02d:%02d", pillName, hour, min)
	msg := tgbotapi.NewMessage(chatId, msgString)
	tb.bot.Send(msg)
	tb.chatStates[chatId] = BotStateNormal
	return nil
}

func (tb TeleBot) addReminder(chatId ChatID) {
	msg := tgbotapi.NewMessage(chatId, "Напишите название препарата:")
	tb.bot.Send(msg)
	tb.chatStates[chatId] = BotStateWaitForInputPillName
}

func RunBot() {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	utls.CheckError(err)

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// Create a timer for sending messages
	messageTimer := time.NewTicker(60 * time.Second)
	defer messageTimer.Stop()

	repo := postgresql.OpenRepo()
	defer repo.CloseRepo()

	chatStates := make(map[ChatID]BotState)
	tmpData := make(map[ChatID]TmpData)

	tb := TeleBot{bot: bot, repo: repo, chatStates: chatStates, tmpData: tmpData}

	for {
		select {
		case update := <-updates:
			if update.CallbackQuery != nil {
				tb.handleCallbackQuery(update.CallbackQuery)
				continue
			}

			if update.Message == nil {
				continue
			}
			err := tb.handleMessage(update)
			if err != nil {
				log.Println(err)
			}

		case t := <-messageTimer.C:
			fmt.Println("Текущее время: ", t)
		}
	}

}
