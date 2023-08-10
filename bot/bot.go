package bot

import (
	"Pills/database/postgresql"
	"Pills/mdl"
	"Pills/utls"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func processMessage(repo mdl.Repo, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {

	isNewcomer, err := repo.UserIsNewcomer(update.Message.Chat.ID)
	if err != nil {
		return err
	}

	if isNewcomer {
		err := repo.SaveUser(
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

	err = repo.SaveMessage(
		update.Message.Chat.ID,
		update.Message.Text,
		update.Message.Time(),
	)
	if err != nil {
		return err
	}

	return nil
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

	// var chatID int64

	// for update := range updates {
	// 	if update.Message == nil {
	// 		continue
	// 	}

	// 	chatID = update.Message.Chat.ID

	// 	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	// 	msg := tgbotapi.NewMessage(chatID, "Hello World!")
	// 	msg.ReplyToMessageID = update.Message.MessageID
	// 	msg.ReplyMarkup = menuKeyboard

	// 	bot.Send(msg)

	// 	err = saveMessageToDB(update)
	// 	log.Println(err)

	// 	// break

	// 	msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Press the button:")
	// 	button := tgbotapi.NewInlineKeyboardButtonData("Callback Button", "callback_data")
	// 	inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
	// 		tgbotapi.NewInlineKeyboardRow(button),
	// 	)
	// 	msg.ReplyMarkup = inlineKeyboard

	// 	bot.Send(msg)
	// }

	repo := postgresql.OpenRepo()
	defer repo.CloseRepo()

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			processMessage(repo, bot, update)

			// log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			// msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello World!")
			// msg.ReplyToMessageID = update.Message.MessageID

			// bot.Send(msg)

		case t := <-messageTimer.C:
			fmt.Println("Текущее время: ", t)
			// msg := tgbotapi.NewMessage(chatID, fmt.Sprintln("Текущее время: ", t))
			// bot.Send(msg)
		}
	}

}
