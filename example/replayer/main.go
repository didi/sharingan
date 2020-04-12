package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/didichuxing/sharingan/replayer"
)

func main() {
	http.HandleFunc("/", indexHandle)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func indexHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!\n")
}
