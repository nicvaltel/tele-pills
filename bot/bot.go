package bot

import (
	"Pills/database/postgresql"
	"Pills/mdl"
	"Pills/utls"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ChatID = int64
type BotState uint8
type ReminderID = int64

const (
	BotStateNormal               BotState = 0
	BotStateWaitForInputPillName BotState = 1
	BotStateWaitForInputHour     BotState = 2
	BotStateWaitForInputMinute   BotState = 3
)

type TeleBot struct {
	bot                 *tgbotapi.BotAPI
	repo                mdl.Repo
	chatStates          map[ChatID]BotState
	pillData            map[ChatID]PillData
	unreadReminders     map[ChatID](map[ReminderID]time.Time)
	delayBetweenReminds time.Duration
}

type PillData struct {
	pillName string
	hour     uint8
	min      uint8
}

type BotFSMStruct struct {
	handleMessage       func(update tgbotapi.Update) error
	handleCallbackQuery func(callback *tgbotapi.CallbackQuery) error
}

func (tb TeleBot) botFSM(botState BotState) BotFSMStruct {
	switch botState {
	case BotStateNormal:
		return BotFSMStruct{
			handleMessage:       tb.handleMessageNormal,
			handleCallbackQuery: tb.handleCallbackQueryNormal,
		}
	case BotStateWaitForInputPillName:
		return BotFSMStruct{
			handleMessage:       tb.handleMessageInputPillName,
			handleCallbackQuery: tb.handleCallbackQueryInputPillName,
		}
	case BotStateWaitForInputHour:
		return BotFSMStruct{
			handleMessage:       tb.handleMessageInputHour,
			handleCallbackQuery: tb.handleCallbackQueryInputHour,
		}
	case BotStateWaitForInputMinute:
		return BotFSMStruct{
			handleMessage:       tb.handleMessageInputMinute,
			handleCallbackQuery: tb.handleCallbackQueryInputMinute,
		}
	default:
		return BotFSMStruct{
			handleMessage:       tb.handleMessageNormal,
			handleCallbackQuery: tb.handleCallbackQueryNormal,
		}
	}
}

func (tb TeleBot) stateToNormal(chatId ChatID) {
	tb.chatStates[chatId] = BotStateNormal
	delete(tb.pillData, chatId)
}

func botLoop(updates tgbotapi.UpdatesChannel, tb TeleBot, messageTimer *time.Ticker) {
	var chatId int64

	for {
		select {
		case update := <-updates:

			if update.Message != nil {
				chatId = update.Message.Chat.ID
			} else if update.CallbackQuery != nil {
				chatId = update.CallbackQuery.Message.Chat.ID
			} else {
				continue
			}

			fsm := tb.botFSM(tb.chatStates[chatId])

			if update.CallbackQuery != nil {
				err := tb.handleAllhandleCallbackQuerys(update.CallbackQuery, fsm.handleCallbackQuery)
				if err != nil {
					log.Println(err)
					tb.stateToNormal(chatId)
				}
				continue
			}

			if update.Message != nil {
				err := tb.handleAllMessages(update, fsm.handleMessage)
				if err != nil {
					log.Println(err)
					tb.stateToNormal(chatId)
				}
			}

		case t := <-messageTimer.C:
			err := tb.reminderRoutine()
			if err != nil {
				log.Println(err)
			}
			log.Println("reminderRoutine at time: ", t)
		}
	}
}

func RunBot() {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	utls.PanicError(err)

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)

	u.Timeout, err = strconv.Atoi(os.Getenv("BOT_TIMEOUT"))
	utls.PanicError(err)

	reminderTicker, err := strconv.Atoi(os.Getenv("BOT_REMINDER_TICKER"))
	utls.PanicError(err)

	delayBetweenRemindsInt, err := strconv.Atoi(os.Getenv("BOT_DELAYS_BETWEEN_REMINDS"))
	utls.PanicError(err)
	delayBetweenReminds := time.Duration(delayBetweenRemindsInt) * time.Second

	updates, err := bot.GetUpdatesChan(u)

	// Create a timer for sending messages
	messageTimer := time.NewTicker(time.Duration(reminderTicker) * time.Second)
	defer messageTimer.Stop()

	repo := postgresql.OpenRepo()
	defer repo.CloseRepo()

	chatStates := make(map[ChatID]BotState)
	pillData := make(map[ChatID]PillData)
	unreadReminders := make(map[ChatID](map[ReminderID]time.Time))

	tb := TeleBot{bot: bot, repo: repo, chatStates: chatStates, pillData: pillData, unreadReminders: unreadReminders, delayBetweenReminds: delayBetweenReminds}

	botLoop(updates, tb, messageTimer)
}
