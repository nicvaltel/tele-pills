package mdl

import (
	"time"
)

// type BotApp interface {
// 	// RunBotApp()
// 	ProcessMessage()
// }

type Repo interface {
	CloseRepo()
	UserIsNewcomer(chatId int64) (bool, error)
	SaveUser(chatId int64, username string, firstName string, lastName string, timestamp time.Time) error
	SaveMessage(chatId int64, message string, timestamp time.Time) error
}
