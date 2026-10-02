package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Task struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Due  string `json:"due"`
	Comp bool   `json:"comp"`
}

func main() {
	var p int = 8089
	var port string = fmt.Sprintf(":%d", p)
	var url string = fmt.Sprintf("http://torment-node%s", port)

	if os.Args[1] == "list" {

	} else if os.Args[1] == "delete" {

	} else if os.Args[1] == "update" {

	} else if os.Args[1] == "genconf" {

	} else if os.Args[1] == "complete" {

	} else if os.Args[1] == "test" {
		w := os.Stdout
		task := Task {Name:"foo", Due:"bar", Id:10, Comp:false}
		err := json.NewEncoder(w).Encode(task)
		if err != nil {
			panic(err)
		}
		_, err = http.Post(url + "/test", "i dunno", w)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("command not recognized, type --help to look at available commands")
	}

}
