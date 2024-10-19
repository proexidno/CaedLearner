package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	botToken := os.Getenv("TELEGRAM_API_TOKEN")

	bot, err := telego.NewBot(botToken, telego.WithWarnings())

	if err != nil {
		fmt.Println("No telegram token")
		os.Exit(1)
	}

	updates, _ := bot.UpdatesViaLongPolling(nil)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := telegoutil.ID(update.Message.Chat.ID)
		bot.CopyMessage(
			telegoutil.CopyMessage(
				chatID,
				chatID,
				update.Message.MessageID,
			),
		)
	}

	defer bot.StopLongPolling()
	return
}
