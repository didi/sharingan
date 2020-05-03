package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/didi/sharingan"
)

func main() {
	http.HandleFunc("/", indexHandle)
	http.HandleFunc("/go", goHandle)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func indexHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!\n")
}

// Tips：正常http请求不需要设置，只有使用go异步执行时需要
func goHandle(w http.ResponseWriter, r *http.Request) {
	doneChan := make(chan bool)

	go func(delegatedID int64) {
		sharingan.SetDelegatedFromGoRoutineID(delegatedID)
		defer sharingan.SetDelegatedFromGoRoutineID(0)
		http.Get("http://127.0.0.1:8888")

		doneChan <- true
	}(sharingan.GetCurrentGoRoutineID())

	<-doneChan
}
