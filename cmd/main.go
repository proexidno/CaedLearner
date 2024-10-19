package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	botToken := os.Getenv("TELEGRAM_API_TOKEN")

	println(botToken)

	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())

	if err != nil {
		fmt.Println("No telegram token")
		os.Exit(1)
	}

	println(bot)
	return
}
