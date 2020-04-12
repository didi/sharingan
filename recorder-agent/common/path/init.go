package path

import (
	"log"
	"os"
)

// Root current dir
var Root string

func init() {
	var err error
	Root, err = os.Getwd()
	if err != nil {
		log.Fatal("Initialize Root error: ", err)
	}
}
