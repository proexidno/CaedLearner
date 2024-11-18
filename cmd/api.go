package main

import (
	"database/sql"
)

type Word struct {
	ID          int    `json:"id"`
	Word        string `json:"word"`
	Translation string `json:"translation"`
	Custom      bool   `json:"custom"`
}

var db *sql.DB

// Which should initialize tables and fill them with words
// error should equal nil unless there is an error when accessing database
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
	CREATE TABLE userwords(  
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		word_id INTEGER,
		next_revise TIME,
		level INTEGER,
		custom BOOLEAN
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}
}

// Func which return a random word from a list of global words that this user have never seen
// error should equal nil unless there is an error when accessing database
func getLearnWord(chatID int64) {

}

// Which should return a random word from a list of words that this user have seen and should revise (determined by time)
// OR nil when there are no words that user can revise
// error should equal nil unless there is an error when accessing database
func getReviseWord(chatID int64) {

}

// Sets word for this user to the next level or revision if isLearned is false
// Sets word for this user to be learned (max level of revision) isLearned is true
// error should equal nil unless there is an error when accessing database
func setWord(word Word, isLearned bool, chatID int64) {

}

func test() {

}
