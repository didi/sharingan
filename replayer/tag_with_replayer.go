// +build replayer

package replayer

import (
	"log"

	"github.com/didichuxing/sharingan/replayer/fastmock"
)

func init() {
	fastmock.MockSyscall()
	fastmock.MockTime()

	log.Println("mode", "=====replayer=====")
}
