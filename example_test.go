package sharingan_test

import (
	"net/http"

	"github.com/didi/sharingan"
)

func Example() {
	doneChan := make(chan bool)

	go func(delegatedID int64) {
		sharingan.SetDelegatedFromGoRoutineID(delegatedID)
		defer sharingan.SetDelegatedFromGoRoutineID(0)

		http.Get("http://127.0.0.1:8888")

		doneChan <- true
	}(sharingan.GetCurrentGoRoutineID())

	<-doneChan
}
