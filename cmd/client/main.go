package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
	"time"
	"go.yaml.in/yaml/v4"
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

type Config struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
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

func list(url string) {
	listTasks, err := getTasks(url)
	if err != nil {
		panic(err)
	}
	nameMax := 0
	dueMax := 0
	keyword := false
	for i := range len(listTasks) {
		if len(listTasks[i].Name) > nameMax {
			nameMax = len(listTasks[i].Name)
		}
		if len(listTasks[i].Due) - 5 > dueMax && !keyword{
			dueMax = len(listTasks[i].Due) -5 
		}
		if listTasks[i].Due == "never" {
			if dueMax < len(listTasks[i].Due) {
				dueMax = len(listTasks[i].Due)
			}
		}
	}
	s := fmt.Sprintf("id | %"+strconv.Itoa(nameMax)+"s | %"+strconv.Itoa(dueMax)+"s | %8s |", "name", "due", "complete")
	fmt.Println(s)
	fmt.Println(strings.Repeat("-", utf8.RuneCountInString(s)))
	var date string
	for i := range len(listTasks) {
		if listTasks[i].Due == "never" {
			date = listTasks[i].Due
		} else {
			dates := strings.Split(listTasks[i].Due, `/`)
			date = dates[0] + "/" + dates[1]
		}
		fmt.Printf("%2d | %"+strconv.Itoa(nameMax)+"s | %"+strconv.Itoa(dueMax)+"s | %8t |\n", listTasks[i].Id, listTasks[i].Name, date, listTasks[i].Comp)
	}
}

func dateCheck(d string) (string, error) {
	if d == "never" {
		return d, nil
	}
	_, err := time.Parse("02/01/2006", d)
	if err != nil {
		return "", err
	}
	
	return d, nil
}

func main() {
	defName := "default"
	defDue := "01/01/1970"
	defId := -1
	name := flag.String("name", defName, "name of task")
	id := flag.Int("id", defId, "task id")
	due := flag.String("due", defDue, "due date of task (DD/MM/YYYY)")
	f := flag.Bool("dev", false, "use ./config.yaml for dev work")

	flag.Parse()
	var err error
	var b []uint8
	if *f {
		b, err = os.ReadFile("config.yaml")
	} else {
		b, err = os.ReadFile("/etc/sharktasks/config.yaml")
	}
	if err != nil {
		panic(err)
	}
	var cfg Config
	err = yaml.Load(b, &cfg)
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	switch os.Args[1] {
	case "list":
		list(url)
	case "insert":
		if *name == defName || *due == defDue {
			if *name == defName {
				fmt.Println("define task name with --name")
			}
			if *due == defDue {
				fmt.Println("define task date with --due")
			}
			break
		}
		var newTask Task
		newTask.Name = *name
		date, err := dateCheck(*due) 
		if err != nil {
			fmt.Println("invalid date, use format DD/MM/YYY")
			break
		} else {
			newTask.Due = date
		}
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(newTask)
		if err != nil {
			panic(err)
		}
		_, err = http.Post(url+"/insert", "application/json", &buf)
		if err != nil {
			panic(err)
		}
		list(url)
	case "delete":
		if *id == -1 {
			fmt.Println("define task id with --id [id]")
			break
		}

		var newTask Task
		newTask.Id = *id
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(newTask)
		if err != nil {
			panic(err)
		}
		resp, err := http.Post(url+"/delete", "application/json", &buf)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if string(body) != "" {
			fmt.Println(string(body))
		} else {
			fmt.Printf("task id %d deleted successfully\n", *id)
			list(url)
		}
	case "complete":
		if *id == -1 {
			fmt.Println("define task id with --id [id]")
			break
		}
		updateTask := Task{Name: defName, Due: defDue, Comp: true, Id: *id}
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(updateTask)
		if err != nil {
			panic(err)
		}
		resp, err := http.Post(url+"/update", "application/json", &buf)
		if err != nil {
			panic(err)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "" {
			fmt.Println(string(body))
		} else {
			fmt.Println("task completed successfully, good boy")
		}
	case "update":
		if *id == -1 {
			fmt.Println("define task id with --id [id]")
			break
		}
		updateTask := Task{Name: *name, Due: *due, Comp: false, Id: *id}
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(updateTask)
		if err != nil {
			panic(err)
		}
		resp, err := http.Post(url+"/update", "application/json", &buf)
		if err != nil {
			panic(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		if string(body) != "" {
			fmt.Println(string(body))
		} else {
			fmt.Printf("task id %d updated successfully\n", id)
			list(url)
		}
	case "del-comp":
		listTasks, err := getTasks(url)
		if err != nil {
			panic(err)
		}
		var delTask Task
		var buf bytes.Buffer
		for i := range slices.Backward(listTasks) {
			if listTasks[i].Comp {
				delTask.Id = listTasks[i].Id
				err = json.NewEncoder(&buf).Encode(delTask)
				if err != nil {
					panic(err)
				}
				_, err = http.Post(url+"/delete", "application/json", &buf)
				if err != nil {
					panic(err)
				}
				buf.Reset()
			}
		}
		list(url)
	case "genconf":
		fmt.Println("---\nhost: \"server url goes here\"\nport: \"server port goes here\"")
	default:
		fmt.Println("command not recognized, type --help to look at available commands")
	}
}
