package main

import (
	"Pills/bot"
	"Pills/database/postgresql"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

var envFile string = "config.env"

func main() {

	errEnv := godotenv.Load(envFile)
	if errEnv != nil {
		log.Panic(fmt.Sprintf("Error loading %s file.\n", envFile))
	}

	postgresql.RunMigration("database/postgresql/migration/migration.sql", "database/postgresql/migration/migration.md5")
	bot.RunBot()

}
