package mdl

import (
	"time"
)

type Reminder struct {
	ChatId      int64 // TODO change from int64 to ChatID and ReminderID
	ReminderId  int64
	ReminderMsg string
}

type Repo interface {
	CloseRepo()
	UserIsNewcomer(chatId int64) (bool, error)
	SaveUser(chatId int64, username string, firstName string, lastName string, timestamp time.Time) error
	SaveMessage(chatId int64, message string, isCallbackQuery bool, timestamp time.Time) error
	SaveReminder(chatId int64, pillName string, hour uint8, min uint8) error
	GetReminds(time time.Time) ([]Reminder, error)
	UpdateRemind(remindId int64) error
}
