package main

import (
	"fmt"
	"os"
	"strings"

	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	// "github.com/proexidno/CardLearner/api"
)

// remove when api is ready
type Word struct {
	ID          int    `json:"id"`
	Word        string `json:"word"`
	Translation string `json:"translation"`
}

type chatState struct {
	word          Word
	whenRequested time.Time
}

var chatStates map[int64]chatState = make(map[int64]chatState)
var chatStateMutex = sync.RWMutex{}

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

func newWordToState(chatID int64) Word {
	chatStateMutex.Lock()
	// random word from db
	var stateWord Word
	var stateState chatState = chatState{word: stateWord, whenRequested: time.Now()}
	chatStates[chatID] = stateState
	chatStateMutex.Unlock()
	return stateWord
}

func callRevise(bot *telego.Bot, update telego.Update) {
	var chatID int64
	if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else {
		chatID = update.CallbackQuery.From.ID
	}

	chatStateMutex.RLock()
	state, ok := chatStates[chatID]
	chatStateMutex.RUnlock()

	var toShow Word
	if !ok {
		toShow = newWordToState(chatID)
	} else {
		if time.Since(state.whenRequested).Minutes() >= 30 {
			toShow = newWordToState(chatID)
		} else {
			toShow = state.word
			state.whenRequested = time.Now()
			chatStateMutex.Lock()
			chatStates[chatID] = state
			chatStateMutex.Unlock()
		}
	}

	println(toShow.Word)

	keyboard := telegoutil.Keyboard(
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("Запомнил(а)"),
			telegoutil.KeyboardButton("Повторить ещё раз"),
		),
	)

	message := telegoutil.Message(
		telegoutil.ID(chatID),
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
