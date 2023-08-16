package postgresql

import (
	"Pills/mdl"
	"database/sql"
	"fmt"
	"time"
)

type PostgresRepo struct {
	db *sql.DB
}

func OpenRepo() mdl.Repo {
	db := openDB()
	return PostgresRepo{db: db}
}

func (repo PostgresRepo) CloseRepo() {
	repo.db.Close()
}

func (repo PostgresRepo) UserIsNewcomer(chatId int64) (bool, error) {
	rows, err := repo.db.Query("SELECT user_id FROM pills.users WHERE user_id = $1", chatId)

	if err != nil {
		return false, err
	}
	return !rows.Next(), nil
}

func (repo PostgresRepo) SaveReminder(chatId int64, pillName string, hour uint8, min uint8) error {
	isNewcomer, err := repo.UserIsNewcomer(chatId)
	if err != nil {
		return err
	}
	if isNewcomer {
		return fmt.Errorf("SaveReminder error: userId %d doesn't exist in Repo", chatId)
	}

	nextRemindTime := nextRemindTime(int(hour), int(min))

	_, err = repo.db.Exec("INSERT INTO pills.pills (user_id, pill_name, pill_hour, pill_min, next_remind_time) VALUES ($1, $2, $3, $4, $5)",
		chatId, pillName, hour, min, nextRemindTime)
	if err != nil {
		return err
	}
	return nil
}

func (repo PostgresRepo) SaveUser(chatId int64, username string, firstName string, lastName string, timestamp time.Time) error {

	isNewcomer, err := repo.UserIsNewcomer(chatId)
	if err != nil {
		return err
	}
	if !isNewcomer {
		return fmt.Errorf("SaveUser error: userId %d already exists in Repo", chatId)
	}

	_, err = repo.db.Exec("INSERT INTO pills.users (user_id, username, first_name, last_name, created) VALUES ($1, $2, $3, $4, $5)",
		chatId, username, firstName, lastName, timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (repo PostgresRepo) SaveMessage(chatId int64, message string, isCallbackQuery bool, timestamp time.Time) error {
	_, err := repo.db.Exec("INSERT INTO pills.messages (user_id, msg_text, is_callback_query, time_sent) VALUES ($1, $2, $3, $4)",
		chatId, message, isCallbackQuery, timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (repo PostgresRepo) GetReminds(t time.Time) ([]mdl.Reminder, error) {
	rows, err := repo.db.Query("SELECT pill_id, user_id, pill_name, next_remind_time FROM pills.pills WHERE next_remind_time < $1", t)
	if err != nil {
		return nil, err
	}

	var pillId int64
	var userId int64
	var pillName string
	var nextTime sql.NullTime

	reminders := make([]mdl.Reminder, 0)
	for rows.Next() {
		rows.Scan(&pillId, &userId, &pillName, &nextTime)
		reminders = append(reminders, mdl.Reminder{
			ChatId:      userId,
			ReminderId:  pillId,
			ReminderMsg: fmt.Sprintf("НАПОМИНАНИЕ: %s в %s", pillName, nextTime.Time.Format(time.ANSIC)),
		})
	}
	return reminders, nil
}

func (repo PostgresRepo) UpdateRemind(remindId int64) error {

	rows, err := repo.db.Query("SELECT pill_hour, pill_min FROM pills.pills WHERE pill_id = $1", remindId)
	if err != nil {
		return err
	}

	rows.Next()
	var hour int
	var min int
	rows.Scan(&hour, &min)

	nextRemindTime := nextRemindTime(hour, min)
	_, err = repo.db.Exec("UPDATE pills.pills set next_remind_time = $1 WHERE pill_id = $2", nextRemindTime, remindId)
	return err
}

func nextRemindTime(hour int, min int) time.Time {
	now := time.Now()
	nextRemindTime := time.Date(now.Year(), now.Month(), now.Day(), int(hour), int(min), 0, 0, now.Location()) // TODO change Location to user location
	if nextRemindTime.Compare(now) < 0 {
		nextRemindTime = nextRemindTime.Add(time.Hour * 24)
	}
	return nextRemindTime
}
