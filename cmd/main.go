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
)

// remove when api is ready
// type Word struct {
// 	ID          int
// 	Word        string
// 	Translation string
// }

type chatState struct {
	word          Word
	whenRequested time.Time
	isLearn       bool
	tries         uint8
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

func newReviseWordToState(chatID int64) Word {
	chatStateMutex.Lock()
	// random word from db
	var stateWord Word = Word{Word: "kekl", Translation: "кекл", ID: 1}
	var stateState chatState = chatState{word: stateWord, whenRequested: time.Now(), isLearn: false}
	chatStates[chatID] = stateState
	chatStateMutex.Unlock()
	return stateWord
}

func newLearnWordToState(chatID int64) Word {
	chatStateMutex.Lock()
	// random word from db
	var stateWord Word = Word{Word: "aboba", Translation: "абоба", ID: 1}
	var stateState chatState = chatState{word: stateWord, whenRequested: time.Now(), isLearn: true}
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
		toShow = newReviseWordToState(chatID)
	} else {
		if time.Since(state.whenRequested).Minutes() >= 30 || state.isLearn {
			toShow = newReviseWordToState(chatID)
		} else {
			toShow = state.word
			state.whenRequested = time.Now()
			chatStateMutex.Lock()
			chatStates[chatID] = state
			chatStateMutex.Unlock()
		}
	}

	fmt.Println(toShow.Word)

	keyboard := telegoutil.Keyboard(
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("Запомнил(а)"),
			telegoutil.KeyboardButton("Повторить ещё раз"),
		),
	)

	message := telegoutil.Message(
		telegoutil.ID(chatID),
		fmt.Sprintf("%v", toShow.Translation),
	).WithReplyMarkup(keyboard)
	bot.SendMessage(message)
}

func callLearn(bot *telego.Bot, update telego.Update) {
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
		toShow = newLearnWordToState(chatID)
	} else {
		if time.Since(state.whenRequested).Minutes() >= 30 || !state.isLearn {
			toShow = newReviseWordToState(chatID)
		} else {
			toShow = state.word
			state.whenRequested = time.Now()
			chatStateMutex.Lock()
			chatStates[chatID] = state
			chatStateMutex.Unlock()
		}
	}

	// fmt.Println(toShow.Translation)

	keyboard := telegoutil.Keyboard(
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("Уже знал(а)"),
			telegoutil.KeyboardButton("Начать учить"),
		),
	)

	message := telegoutil.Message(
		telegoutil.ID(chatID),
		fmt.Sprintf("%v", toShow.Translation),
	).WithReplyMarkup(keyboard)
	bot.SendMessage(message)
}

func handleUpdate(bot *telego.Bot, update telego.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	raw := update.Message.Text
	concated := strings.TrimSpace(raw)
	userMessage := strings.ToLower(concated)

	chatStateMutex.RLock()
	state, ok := chatStates[chatID]
	chatStateMutex.RUnlock()

	if userMessage == "запомнил" || userMessage == "запомнила" || userMessage == "запомнил(а)" {
		chatStateMutex.Lock()
		currState := chatStates[chatID]
		delete(chatStates, chatID)
		chatStateMutex.Unlock()
		// call db next level for currState
		fmt.Println(currState.word)
		callRevise(bot, update)
		return
	} else if userMessage == "уже знал" || userMessage == "уже знала" || userMessage == "уже знал(а)" {
		chatStateMutex.Lock()
		currState := chatStates[chatID]
		delete(chatStates, chatID)
		chatStateMutex.Unlock()
		// call db max level for currState
		fmt.Println(currState.word)
		callLearn(bot, update)
		return
	} else if userMessage == "начать учить" {
		chatStateMutex.Lock()
		currState := chatStates[chatID]
		delete(chatStates, chatID)
		chatStateMutex.Unlock()
		// call db start revising new word
		fmt.Println(currState.word)
		callLearn(bot, update)
		return
	} else if userMessage == "повторить ещё раз" || userMessage == "повторить еще раз" {
		chatStateMutex.Lock()
		delete(chatStates, chatID)
		chatStateMutex.Unlock()
		callRevise(bot, update)
		return
	}

	if !ok {
		callMenu(bot, update)
		return
	}
	var message *telego.SendMessageParams

	fmt.Println(userMessage)
	if userMessage != state.word.Word && state.tries < 2 {
		state.tries++
		chatStateMutex.Lock()
		chatStates[chatID] = state
		chatStateMutex.Unlock()

		// wrong tries: x/3
		message = telegoutil.Message(
			telegoutil.ID(chatID),

			fmt.Sprintf("%v\nНеправильно, попыток: %v/3", state.word.Word, state.tries),
		)

	} else {
		var keyboard *telego.ReplyKeyboardMarkup
		if state.isLearn {
			keyboard = telegoutil.Keyboard(
				telegoutil.KeyboardRow(
					telegoutil.KeyboardButton("Уже знал(а)"),
					telegoutil.KeyboardButton("Начать учить"),
				),
			)
		} else {
			keyboard = telegoutil.Keyboard(
				telegoutil.KeyboardRow(
					telegoutil.KeyboardButton("Запомнил(а)"),
					telegoutil.KeyboardButton("Повторить ещё раз"),
				),
			)
		}
		message = telegoutil.Message(
			telegoutil.ID(chatID),
			fmt.Sprintf("%v\n%v", state.word.Word, state.word.Translation),
		).WithReplyMarkup(keyboard)
	}

	bot.SendMessage(message)
}

func main() {
	const debug bool = true
	if debug {
		test()
		os.Exit(0)
	}
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
	botHandler.Handle(handleUpdate, telegohandler.AnyMessageWithText())

	println("Up and running")
	botHandler.Start()

	return
}
