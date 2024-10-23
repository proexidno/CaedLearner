package main

import (
	"fmt"
	"os"
	"strings"

	// "sync"
	// "time"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	// "github.com/proexidno/CardLearner/api"
)

// type chatState struct {
// 	word          api.Word
// 	whenRequested time.Time
// }

// var chatStates map[string]chatState = make(map[string]chatState)
// var chatStateMutex = sync.Mutex{}

func callMenu(bot *telego.Bot, update telego.Update) {
	buttonRevise := telegoutil.InlineKeyboardButton("Повторять слова")
	buttonRevise.CallbackData = "revise"

	buttonLearn := telegoutil.InlineKeyboardButton("Учить новые слова")
	buttonLearn.CallbackData = "learn"

	keyboard := telegoutil.InlineKeyboard(
		telegoutil.InlineKeyboardRow(
			buttonRevise,
			buttonLearn,
		),
	)
	message := telegoutil.Message(
		telegoutil.ID(update.Message.Chat.ID),
		"This is menu",
	).WithReplyMarkup(keyboard)
	_, err := bot.SendMessage(message)
	if err != nil {
		println(err.Error())
	}
}

func callRevise(bot *telego.Bot, update telego.Update) {
	var chatID telego.ChatID
	if update.Message != nil {
		chatID = telegoutil.ID(update.Message.Chat.ID)
	} else {
		chatID = telegoutil.ID(update.CallbackQuery.From.ID)
	}

	keyboard := telegoutil.Keyboard(
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("Запомнил(а)"),
			telegoutil.KeyboardButton("Повторить ещё раз"),
		),
	)

	message := telegoutil.Message(
		chatID,
		"This is revise",
	).WithReplyMarkup(keyboard)
	bot.SendMessage(message)
}

func callLearn(bot *telego.Bot, update telego.Update) {
	var chatID telego.ChatID
	if update.Message != nil {
		chatID = telegoutil.ID(update.Message.Chat.ID)
	} else {
		chatID = telegoutil.ID(update.CallbackQuery.From.ID)
	}
	message := telegoutil.Message(
		chatID,
		"This is learn",
	)
	bot.SendMessage(message)
}

func handleQuery(bot *telego.Bot, update telego.Update) {
	fmt.Println("query:", update.CallbackQuery.Data)
}

func hadleUpdate(bot *telego.Bot, update telego.Update) {
	if update.Message == nil {
		return
	}

	chatID := telegoutil.ID(update.Message.Chat.ID)

	raw := update.Message.Text
	concated := strings.TrimSpace(raw)
	userMessage := strings.ToLower(concated)
	println(userMessage)

	bot.CopyMessage(
		telegoutil.CopyMessage(
			chatID,
			chatID,
			update.Message.MessageID,
		),
	)
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
	botHandler, _ := telegohandler.NewBotHandler(bot, updates)

	defer bot.StopLongPolling()
	defer botHandler.Stop()

	botHandler.Handle(callMenu, telegohandler.CommandEqual("start"))
	botHandler.Handle(callRevise, telegohandler.CallbackDataEqual("revise"))
	botHandler.Handle(callLearn, telegohandler.CallbackDataEqual("learn"))
	botHandler.Handle(callRevise, telegohandler.CommandEqual("revise"))
	botHandler.Handle(callLearn, telegohandler.CommandEqual("learn"))
	botHandler.Handle(hadleUpdate, telegohandler.AnyMessage())

	println("Up and running")
	botHandler.Start()

	return
}
