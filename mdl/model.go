package mdl

import (
	"time"
)

// type BotApp interface {
// 	ProcessMessage(repo Repo, update tgbotapi.Update) error
// 	sendAddPillButton()
// }

type Repo interface {
	CloseRepo()
	UserIsNewcomer(chatId int64) (bool, error)
	SaveUser(chatId int64, username string, firstName string, lastName string, timestamp time.Time) error
	SaveMessage(chatId int64, message string, timestamp time.Time) error
	SaveReminder(chatId int64, pillName string, hour uint8, min uint8) error
}
