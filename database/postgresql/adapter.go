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

	_, err = repo.db.Exec("INSERT INTO pills.pills (user_id, pill_name, pill_hour, pill_min) VALUES ($1, $2, $3, $4)",
		chatId, pillName, hour, min)
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

func (repo PostgresRepo) SaveMessage(chatId int64, message string, timestamp time.Time) error {
	_, err := repo.db.Exec("INSERT INTO pills.messages (user_id, msg_text, time_sent) VALUES ($1, $2, $3)",
		chatId, message, timestamp)
	if err != nil {
		return err
	}
	return nil
}
