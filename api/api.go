package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
)

type Word struct {
	ID          int    `json:"id"`
	Word        string `json:"word"`
	Translation string `json:"translation"`
}

var db *sql.DB

func initializeDatabase() {
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
	CREATE TABLE IF NOT EXISTS userwords (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		word_id INTEGER,
		next_revixe TIME,
		custom BOOLEAN
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		panic(err)
	}
}

func createWord(c *gin.Context) {
	var newWord Word
	if err := c.ShouldBindJSON(&newWord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingWord Word
	err := db.QueryRow("SELECT id, word, translation FROM words WHERE word = ?", newWord.Word).Scan(&existingWord.ID, &existingWord.Word, &existingWord.Translation)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Word already exists"})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to check word uniqueness"})
		return
	}

	_, err = db.Exec("INSERT INTO words (word, translation) VALUES (?, ?)", newWord.Word, newWord.Translation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create word"})
		return
	}

	c.JSON(http.StatusCreated, newWord)
}

func getWords(c *gin.Context) {
	rows, err := db.Query("SELECT id, word, translation FROM words ORDER BY RANDOM() LIMIT 5")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch words"})
		return
	}
	defer rows.Close()

	var words []Word
	for rows.Next() {
		var word Word
		if err := rows.Scan(&word.ID, &word.Word, &word.Translation); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to scan word"})
			return
		}
		words = append(words, word)
	}

	c.JSON(http.StatusOK, words)
}

func test() {
	initializeDatabase()
	defer db.Close()

	r := gin.Default()

	r.GET("/words", getWords)
	r.POST("/words", createWord)

	r.Run()
}
