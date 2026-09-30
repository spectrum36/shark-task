package main
	
import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"fmt"
	"encoding/json"
)

type Task struct {
	Id int
	Name string
	Due string
	Comp bool
}

func main() {
	db, err := sql.Open("sqlite", "./tasks.db")
	if err != nil {
		panic(err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS task (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name STRING NOT NULL,
		due STRING,
		complete BOOLEAN NOT NULL DEFAULT(0)
		)`)

	http.HandleFunc("/insert", insert(db))
	http.HandleFunc("/delete", delete(db))
	fmt.Println("server up")
	log.Fatal(http.ListenAndServe(":8089", nil))
}

func insert(db *sql.DB) http.HandlerFunc { 
	return func (w http.ResponseWriter, r *http.Request) {
		var newTask Task

		err := json.NewDecoder(r.Body).Decode(&newTask)
		if err != nil {
			panic(err)
		}
		query := fmt.Sprintf("INSERT INTO task (name, due) VALUES ('%s', '%s');", newTask.Name, newTask.Due)
		fmt.Println(query)
		db.Exec(query)
	}
}

func delete(db *sql.DB) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		var deleteTask Task
		err := json.NewDecoder(r.Body).Decode(&deleteTask)
		if err != nil {
			panic(err)
		}
		query := fmt.Sprintf("DELETE FROM task WHERE name = '%s';", deleteTask.Name)
		db.Exec(query)
	}
}
