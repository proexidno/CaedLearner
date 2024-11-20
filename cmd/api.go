package main

import (
	"database/sql"
	"encoding/csv"
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
func initDatabase() {
	var err error
	db, err = sql.Open("sqlite3", "./data/database.db")
	if err != nil {
		panic(err)
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
		custom BOOLEAN
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}

	loadWordsFromCSV("data/words1.csv")
	loadWordsFromCSV("data/words2.csv")
	loadWordsFromCSV("data/words3.csv")
}

// Load words from a CSV file into the database
func loadWordsFromCSV(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Error opening CSV file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Println("Error reading CSV file:", err)
		return
	}

	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		_, err := db.Exec("INSERT OR IGNORE INTO words (word, translation, custom) VALUES (?, ?, ?)", record[0], record[1], false)
		if err != nil {
			log.Println("Error inserting word into DB:", err)
		}
	}
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
		FROM words w
		JOIN userwords uw ON w.id = uw.word_id
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

// Set a word for the user to the next level or mark it as learned
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
			INSERT OR REPLACE INTO userwords (user_id, word_id, next_revise, level, custom) 
			VALUES (?, ?, ?, 0, ?)`, chatID, word.ID, nextRevise, word.Custom)
		return err
	}
}

func test() {
	initDatabase()
}
