package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func hadleUpdate(bot *telego.Bot, update telego.Update) {
	if update.Message == nil {
		return
	}

	chatID := telegoutil.ID(update.Message.Chat.ID)

	keyboard := telegoutil.Keyboard(
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("Start"),
		),
	)

	message := telegoutil.Message(
		chatID,
		"Keyboard message",
	).WithReplyMarkup(keyboard)

	bot.SendMessage(message)
}

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
		hadleUpdate(bot, update)
	}

	defer bot.StopLongPolling()
	return
}
