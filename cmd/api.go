package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Word struct {
	ID          int    `json:"id"`
	Word        string `json:"word"`
	Translation string `json:"translation"`
	Custom      bool   `json:"custom"`
}

var db *sql.DB

// Initialize the database and populate it with words from CSV files
func initDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		println(err.Error())
		return err
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS words (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT NOT NULL UNIQUE,
		translation TEXT NOT NULL,
		custom BOOLEAN
	);
	CREATE TABLE IF NOT EXISTS userwords(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		word_id INTEGER,
		next_revise TIMESTAMP,
		level INTEGER,
		custom BOOLEAN,
		UNIQUE(user_id, word_id) -- Уникальное ограничение на сочетание user_id и word_id
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		println(err.Error())
		return err
	}

	err = loadWordsFromCSV("data/words1.csv")
	if err != nil {
		return err
	}
	err = loadWordsFromCSV("data/words2.csv")
	if err != nil {
		return err
	}
	err = loadWordsFromCSV("data/words3.csv")
	return err
}

func openDataBase() {
	var err error
	db, err = sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		log.Fatalf("Error when opening db:\n%v\n", err.Error())
	}
}

// Load words from a CSV file into the database
func loadWordsFromCSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Error opening CSV file: %v\n", err.Error())
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV file: %v\n", err.Error())
		return err
	}

	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		_, err := db.Exec("INSERT OR IGNORE INTO words (word, translation, custom) VALUES (?, ?, ?)", record[0], record[1], false)
		if err != nil {
			log.Printf("Error inserting word into DB: %v\n", err.Error())
		}
	}
	return nil
}

// Get a random word that the user has never seen
func getLearnWord(chatID int64) (*Word, error) {
	var word Word
	row := db.QueryRow(`
		SELECT id, word, translation, custom 
		FROM words 
		WHERE id NOT IN (SELECT word_id FROM userwords WHERE user_id = ?)
		ORDER BY RANDOM() LIMIT 1`, chatID)

	if err := row.Scan(&word.ID, &word.Word, &word.Translation, &word.Custom); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &word, nil
}

// Get a random word that the user should revise
func getReviseWord(chatID int64) (*Word, error) {
	var word Word
	row := db.QueryRow(`
		SELECT w.id, w.word, w.translation, w.custom 
		FROM words w INNER JOIN userwords uw 
		ON uw.word_id = w.id
		WHERE uw.user_id = ? AND uw.next_revise <= ?
		ORDER BY RANDOM() LIMIT 1`, chatID, time.Now())

	if err := row.Scan(&word.ID, &word.Word, &word.Translation, &word.Custom); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &word, nil
}

// Sets word for this user to the next level or revision if isLearned is false
// Sets word for this user to be learned (max level of revision) isLearned is true
func setWord(word Word, isLearned bool, chatID int64) error {
	now := time.Now()
	if isLearned {
		_, err := db.Exec(`
			UPDATE userwords 
			SET level = level + 1, next_revise = NULL 
			WHERE user_id = ? AND word_id = ?`, chatID, word.ID)
		return err
	} else {
		nextRevise := now.Add(24 * time.Hour)
		_, err := db.Exec(`
			INSERT INTO userwords (user_id, word_id, next_revise, level, custom) 
			VALUES (?, ?, ?, 0, ?) 
			ON CONFLICT (user_id, word_id) 
			DO UPDATE SET next_revise = ?, level = 0`,
			chatID, word.ID, nextRevise, word.Custom, nextRevise)
		return err
	}
}

func test() {
	err := initDatabase()
	if err != nil {
		log.Fatalln("error Initing db")
	}

	chatID := int64(1)

	word, err := getLearnWord(chatID)
	if err != nil {
		log.Fatalf("Ошибка при получении слова для изучения: %v", err)
	} else if word == nil {
		fmt.Println("Нет доступных слов для изучения.")
	} else {
		fmt.Printf("Слово для изучения: %s - Перевод: %s\n", word.Word, word.Translation)
		err := setWord(*word, false, chatID)
		if err != nil {
			log.Fatalf("Ошибка при обновлении статуса слова: %v", err)
		}
	}

	reviseWord, err := getReviseWord(chatID)
	if err != nil {
		log.Fatalf("Ошибка при получении слова для повторения: %v", err)
	} else if reviseWord == nil {
		fmt.Println("Нет доступных слов для повторения.")
	} else {
		fmt.Printf("Слово для повторения: %s - Перевод: %s\n", reviseWord.Word, reviseWord.Translation)
		err := setWord(*reviseWord, true, chatID)
		if err != nil {
			log.Fatalf("Ошибка при обновлении статуса слова: %v", err)
		}
	}

	if word, err := getLearnWord(chatID); err == nil && word != nil {
		fmt.Printf("Новое слово для изучения: %s - Перевод: %s\n", word.Word, word.Translation)
	}

	if reviseWord, err := getReviseWord(chatID); err == nil && reviseWord != nil {
		fmt.Printf("Новое слово для повторения: %s - Перевод: %s\n", reviseWord.Word, reviseWord.Translation)
	}
}
