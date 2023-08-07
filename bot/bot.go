package bot

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func RunHelloWorldBot() {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	// Create a timer for sending messages
	messageTimer := time.NewTicker(2 * time.Second)
	defer messageTimer.Stop()

	var chatID int64

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID = update.Message.Chat.ID

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(chatID, "Hello World!")
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
		break
	}

	for {
		select {
		case update := <-updates:
			if update.Message == nil {
				continue
			}

			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello World!")
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)

		case t := <-messageTimer.C:
			// fmt.Println("Текущее время: ", t)
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintln("Текущее время: ", t))
			bot.Send(msg)
		}
	}
}

func AAA() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		time.Sleep(10 * time.Second)
		done <- true
	}()
	for {
		select {
		case <-done:
			fmt.Println("Готово!")
			return
		case t := <-ticker.C:
			fmt.Println("Текущее время: ", t)
		}
	}
}
