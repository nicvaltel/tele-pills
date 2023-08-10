package main

import (
	"Pills/bot"
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

	// postgresql.RunMigration()
	bot.RunBot()
}
