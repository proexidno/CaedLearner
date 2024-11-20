package main

import (
	"fmt"
	"log"
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
		`Чтобы вызвать это меню напишите /start или /menu
Для начала повторения напишите /revise или нажмите на кнопку ниже
Для заучивания новых слов напишите /learn или нажмите кнопку ниже

Повторения карточек в этом боте основано на методе повторяющегося обучения
Если подтвердить обучение карточки будут появляться в повторении через увеличивающийся промежутки времени`,
	).WithReplyMarkup(keyboard)
	bot.SendMessage(message)
}

func newReviseWordToState(chatID int64) (*Word, error) {
	chatStateMutex.Lock()
	defer chatStateMutex.Unlock()
	// random word from db
	stateWord, err := getReviseWord(chatID)
	if err != nil {
		return nil, err
	}
	if stateWord == nil {
		return nil, nil
	}
	var stateState chatState = chatState{word: *stateWord, whenRequested: time.Now(), isLearn: false}
	chatStates[chatID] = stateState
	return stateWord, nil
}

func newLearnWordToState(chatID int64) (*Word, error) {
	chatStateMutex.Lock()
	defer chatStateMutex.Unlock()
	// random word from db
	stateWord, err := getLearnWord(chatID)
	if err != nil {
		return nil, err
	}
	var stateState chatState = chatState{word: *stateWord, whenRequested: time.Now(), isLearn: true}
	chatStates[chatID] = stateState
	return stateWord, nil
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

	var toShow *Word
	var err error
	if !ok {
		toShow, err = newReviseWordToState(chatID)
		if err != nil {
			log.Printf("error when newReviseWord:\n%v\n", err.Error())
			return
		}
		if toShow == nil {
			toShow, err = newLearnWordToState(chatID)
			if err != nil {
				log.Printf("error when newLearnWord:\n%v\n", err.Error())
				return
			}
			keyboard := telegoutil.Keyboard(
				telegoutil.KeyboardRow(
					telegoutil.KeyboardButton("Запомнил(а)"),
					telegoutil.KeyboardButton("Повторить ещё раз"),
				),
			)

			message := telegoutil.Message(
				telegoutil.ID(chatID),
				fmt.Sprintf("Нет доступных слов для повторения. Вот новое слово:\n%v", (*toShow).Translation),
			).WithReplyMarkup(keyboard)
			bot.SendMessage(message)
			return
		}
	} else {
		if time.Since(state.whenRequested).Minutes() >= 30 || state.isLearn {
			toShow, err = newReviseWordToState(chatID)
			if err != nil {
				log.Printf("error when newReviseWord:\n%v\n", err.Error())
				return
			}
			if toShow == nil {
				toShow, err = newLearnWordToState(chatID)
				if err != nil {
					log.Printf("error when newLearnWord:\n%v\n", err.Error())
					return
				}
				keyboard := telegoutil.Keyboard(
					telegoutil.KeyboardRow(
						telegoutil.KeyboardButton("Запомнил(а)"),
						telegoutil.KeyboardButton("Повторить ещё раз"),
					),
				)

				message := telegoutil.Message(
					telegoutil.ID(chatID),
					fmt.Sprintf("Нет доступных слов для повторения. Вот новое слово:\n%v", (*toShow).Translation),
				).WithReplyMarkup(keyboard)
				bot.SendMessage(message)
			}
		} else {
			toShow = &state.word
			state.whenRequested = time.Now()
			chatStateMutex.Lock()
			chatStates[chatID] = state
			chatStateMutex.Unlock()
		}
	}

	if toShow == nil {
		message := telegoutil.Message(
			telegoutil.ID(chatID),
			"Что-то пошло не так",
		)
		bot.SendMessage(message)
		return
	}

	keyboard := telegoutil.Keyboard(
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("Запомнил(а)"),
			telegoutil.KeyboardButton("Повторить ещё раз"),
		),
	)

	message := telegoutil.Message(
		telegoutil.ID(chatID),
		fmt.Sprintf("%v", (*toShow).Translation),
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

	var toShow *Word
	var err error
	if !ok {
		toShow, err = newLearnWordToState(chatID)
		if err != nil {
			log.Printf("error when newReviseWord:\n%v\n", err.Error())
			return
		}
		if toShow == nil {
			toShow, err = newLearnWordToState(chatID)
			if err != nil {
				log.Printf("error when newLearnWord:\n%v\n", err.Error())
				return
			}
		}
	} else {
		if time.Since(state.whenRequested).Minutes() >= 30 || !state.isLearn {
			toShow, err = newReviseWordToState(chatID)
			if err != nil {
				log.Printf("error when newLearnWord:\n%v\n", err.Error())
				return
			}
			if toShow == nil {
				toShow, err = newLearnWordToState(chatID)
				if err != nil {
					log.Printf("error when newLearnWord:\n%v\n", err.Error())
					return
				}
			}
		} else {
			toShow = &state.word
			state.whenRequested = time.Now()
			chatStateMutex.Lock()
			chatStates[chatID] = state
			chatStateMutex.Unlock()
		}
	}

	if toShow == nil {
		message := telegoutil.Message(
			telegoutil.ID(chatID),
			"Что-то пошло не так",
		)
		bot.SendMessage(message)
		return
	}

	keyboard := telegoutil.Keyboard(
		telegoutil.KeyboardRow(
			telegoutil.KeyboardButton("Уже знал(а)"),
			telegoutil.KeyboardButton("Начать учить"),
		),
	)

	message := telegoutil.Message(
		telegoutil.ID(chatID),
		fmt.Sprintf("%v", (*toShow).Translation),
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
		// call db next level for currState
		setWord(currState.word, false, chatID)
		chatStateMutex.Unlock()
		callRevise(bot, update)
		return
	} else if userMessage == "уже знал" || userMessage == "уже знала" || userMessage == "уже знал(а)" {
		chatStateMutex.Lock()
		currState := chatStates[chatID]
		delete(chatStates, chatID)
		// call db max level for currState
		setWord(currState.word, true, chatID)
		chatStateMutex.Unlock()
		callLearn(bot, update)
		return
	} else if userMessage == "начать учить" {
		chatStateMutex.Lock()
		currState := chatStates[chatID]
		delete(chatStates, chatID)
		// call db start revising new word
		setWord(currState.word, false, chatID)
		chatStateMutex.Unlock()
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

	if userMessage != state.word.Word && state.tries < 2 {
		state.tries++
		chatStateMutex.Lock()
		chatStates[chatID] = state
		chatStateMutex.Unlock()

		// wrong tries: x/3
		message = telegoutil.Message(
			telegoutil.ID(chatID),
			fmt.Sprintf("%v\nНеправильно, попыток: %v/3", state.word.Translation, state.tries),
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
	const debug bool = false
	if debug {
		log.Println("Test up and running")
		test()
		os.Exit(0)
	}

	err := godotenv.Load()

	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	botToken := os.Getenv("TELEGRAM_API_TOKEN")

	isDocker := os.Getenv("DOCKER")

	var datadir string
	if isDocker == "true" {
		datadir = "/data"
	} else {
		datadir = "./data"
	}

	_, err = os.Stat(datadir + "/database.db")
	if err != nil {
		log.Printf("Creating database\n")
		err = initDatabase(datadir)
		if err != nil {
			log.Fatalf("Error when Initing database:\n%v\n", err.Error())
		}
	}
	openDataBase(datadir)

	bot, err := telego.NewBot(botToken, telego.WithWarnings())

	if err != nil {
		log.Fatalln("No telegram token")
	}

	updates, _ := bot.UpdatesViaLongPolling(nil)
	botHandler, _ := telegohandler.NewBotHandler(bot, updates)

	defer bot.StopLongPolling()
	defer botHandler.Stop()

	botHandler.Handle(callMenu, telegohandler.CommandEqual("start"))
	botHandler.Handle(callMenu, telegohandler.CommandEqual("menu"))
	botHandler.Handle(callRevise, telegohandler.CallbackDataEqual("revise"))
	botHandler.Handle(callLearn, telegohandler.CallbackDataEqual("learn"))
	botHandler.Handle(callRevise, telegohandler.CommandEqual("revise"))
	botHandler.Handle(callLearn, telegohandler.CommandEqual("learn"))
	botHandler.Handle(handleUpdate, telegohandler.AnyMessageWithText())

	log.Println("Up and running")
	botHandler.Start()

	return
}
