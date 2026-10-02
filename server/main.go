package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
)

type Task struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Due  string `json:"due"`
	Comp bool   `json:"comp"`
}

func readTask(r *http.Request) (Task, error) {
	var newTask Task
	err := json.NewDecoder(r.Body).Decode(&newTask)
	if err != nil {
		return newTask, err
	}
	return newTask, nil
}

func writeTask(w http.ResponseWriter, task []Task) error{
	err := json.NewEncoder(w).Encode(task)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	db, err := sql.Open("sqlite", "./tasks.db")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS task (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name STRING NOT NULL,
		due STRING,
		complete BOOLEAN NOT NULL DEFAULT(0)
		)`)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/insert", insert(db))
	http.HandleFunc("/delete", delete(db))
	http.HandleFunc("/update", update(db))
	http.HandleFunc("/list", list(db))
	http.HandleFunc("/test", test)
	fmt.Println("server up")
	log.Fatal(http.ListenAndServe(":8089", nil))
}

func insert(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newTask, err := readTask(r)
		if err != nil {
			panic(err)
		}
		query := fmt.Sprintf("INSERT INTO task (name, due) VALUES ('%s', '%s');", newTask.Name, newTask.Due)
		fmt.Println(query)
		_, err = db.Exec(query)
		if err != nil {
			panic(err)
		}

	}
}

func delete(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deleteTask, err := readTask(r)
		if err != nil {
			panic(err)
		}
		query := fmt.Sprintf("DELETE FROM task WHERE id = '%d';", deleteTask.Id)
		_, err = db.Exec(query)
		if err != nil {
			panic(err)
		}

	}
}

func update(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateTask, err := readTask(r)
		if err != nil {
			panic(err)
		}
		query := fmt.Sprintf("UPDATE task SET name = '%s', due = '%s', complete = '%t' WHERE id = %d;", updateTask.Name, updateTask.Due, updateTask.Comp, updateTask.Id)

		_, err = db.Exec(query)
		if err != nil {
			panic(err)
		}
	}
}

func list(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := "SELECT * FROM task ORDER BY id"
		rows, err := db.Query(query)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		err = rows.Err()
		if err != nil {
			panic(err)
		}
		var listTasks []Task
		
		for rows.Next() {
			t := &Task{}
			err = rows.Scan(&t.Id, &t.Comp, &t.Due, &t.Name)
			if err != nil {
				panic(err)
			}
			listTasks = append(listTasks, *t)
		}
		err = writeTask(w, listTasks)
		if err != nil {
			panic(err)
		}
	}
}

func test(w http.ResponseWriter, r *http.Request) {
	newTask, err := readTask(r)
	if err != nil {
		panic(err)
	}
	fmt.Println(newTask)
}
