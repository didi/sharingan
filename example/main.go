package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/didichuxing/sharingan/recorder"
	_ "github.com/didichuxing/sharingan/replayer"
	// TODO：最后import其他业务包！
)

func main() {
	http.HandleFunc("/", indexHandle)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func indexHandle(w http.ResponseWriter, r *http.Request) {
	testHTTPRequest()
	fmt.Fprintf(w, "Hello world!\n")
}

func testHTTPRequest() {
	// 如有需要本地启动一个8888端口提供服务
	rsp, err := http.Get("http://127.0.0.1:8888")
	if err != nil {
		fmt.Printf("[testHTTPRequest][err:%v]\n", err)
		return
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		fmt.Println("res:", string(body), err)
	}
}
