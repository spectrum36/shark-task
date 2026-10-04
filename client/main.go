package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	flag "github.com/spf13/pflag"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

/*
todo:
	add date checker
	make new go file for notifications
	add .yaml supoort
	make genconf do something
*/

type Task struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Due  string `json:"due"`
	Comp bool   `json:"comp"`
}

func getTasks(url string) ([]Task, error) {
	var listTasks []Task
	resp, err := http.Get(url + "/list")
	if err != nil {
		return listTasks, err
	}
	err = json.NewDecoder(resp.Body).Decode(&listTasks)
	if err != nil {
		return listTasks, err
	}
	return listTasks, err
}

func main() {
	defName := "default"
	defDue := "01/01/1970"
	defId := -1
	var name = flag.String("name", defName, "name of task")
	var id = flag.Int("id", defId, "task id")
	var due = flag.String("due", defDue, "due date of task")

	flag.Parse()


	p := 8089
	port := fmt.Sprintf(":%d", p)
	url := fmt.Sprintf("http://torment-node%s", port)

	switch os.Args[1] {
	case "list":
		listTasks, err := getTasks(url)
		if err != nil {
			panic(err)
		}
		nameMax := 0
		dueMax := 0
		for i := 0; i < len(listTasks); i++ {
			if len(listTasks[i].Name) > nameMax {
				nameMax = len(listTasks[i].Name)
			}
			if len(listTasks[i].Due) > dueMax {
				dueMax = len(listTasks[i].Due)
			}
		}
		s := fmt.Sprintf("id | %"+strconv.Itoa(nameMax)+"s | %"+strconv.Itoa(dueMax)+"s | %8s |", "name", "due", "complete")
		fmt.Println(s)
		fmt.Println(strings.Repeat("-", utf8.RuneCountInString(s)))
		for i := 0; i < len(listTasks); i++ {
			fmt.Printf("%2d | %"+strconv.Itoa(nameMax)+"s | %"+strconv.Itoa(dueMax)+"s | %8t |\n", listTasks[i].Id, listTasks[i].Name, listTasks[i].Due, listTasks[i].Comp)
		}
	case "insert":
		var newTask Task
		newTask.Name = *name
		newTask.Due = *due
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(newTask)
		if err != nil {
			panic(err)
		}
		_, err = http.Post(url+"/insert", "application/json", &buf)
		if err != nil {
			panic(err)
		}
	case "delete":
		var newTask Task
		newTask.Id = *id
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(newTask)
		if err != nil {
			panic(err)
		}
		_, err = http.Post(url+"/delete", "application/json", &buf)
		if err != nil {
			panic(err)
		}
	case "complete":
		if *id == -1 {
			fmt.Println("define task id with --id [id]")
			break
		}
		updateTask := Task{Name:defName, Due:defDue, Comp:true, Id:*id}
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(updateTask)
		if err != nil {
			panic(err)
		}
		_, err = http.Post(url + "/update", "application/json", &buf)
		if err != nil {
			panic(err)
		}
		fmt.Println("task completed successfully, good boy")
		
	case "update":
		if *id == -1 {
			fmt.Println("define task id with --id [id]")
			break
		}
		updateTask := Task{Name:*name, Due:*due, Comp:false, Id:*id}
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(updateTask)
		if err != nil {
			panic(err)
		}
		_, err = http.Post(url + "/update", "application/json", &buf)
		if err != nil {
			panic(err)
		}	
	case "test":
		var buf bytes.Buffer
		task := Task{Name: "foo", Due: "bar", Id: 10, Comp: false}
		err := json.NewEncoder(&buf).Encode(task)
		if err != nil {
			panic(err)
		}
		_, err = http.Post(url+"/test", "i dunno", &buf)
		if err != nil {
			panic(err)
		}
	case "del-comp":
		listTasks, err := getTasks(url)
		if err != nil {
			panic(err)
		}
		var delTask Task
		var buf bytes.Buffer
		for i := len(listTasks) - 1; i >= 0; i-- {
			if listTasks[i].Comp {
				delTask.Id = listTasks[i].Id
				err = json.NewEncoder(&buf).Encode(delTask)
				if err != nil {
					panic(err)
				}
				_, err = http.Post(url + "/delete", "application/json", &buf)
				if err != nil {
					panic(err)
				}
				buf.Reset()
			}
		}
		
	case "genconf":

	default:
		fmt.Println("command not recognized, type --help to look at available commands")
	}
}
