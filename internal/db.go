package internal

// database package that stores content to sqllite
import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	Id        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Completed bool      `json:"completed"`
}

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// DB is the database connection
var DB *sql.DB

// InitDB initializes the database
func InitDB() {
	if _, err := os.Stat("./todos.db"); os.IsNotExist(err) {
		log.Println("Creating todos table")
		file, err := os.Create("./todos.db")
		if err != nil {
			log.Fatal("Error creating database file:", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite3", "./todos.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	initMigration, err := os.ReadFile("./migrations/init.sql")
	if err != nil {
		log.Fatal("Error reading init migration:", err)
	}

	_, err = db.Exec(string(initMigration))
	if err != nil {
		log.Fatal("Error creating todos table:", err)
	}

	DB = db
}
