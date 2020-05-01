// +build recorder replayer

package sharingan_test

import (
	"fmt"

	"github.com/didi/sharingan"
)

func Example() {
	sharingan.SetDelegatedFromGoRoutineID(1)
	goroutineID := sharingan.GetCurrentGoRoutineID()

	//Output:
	//goroutineID = 1
	fmt.Printf("goroutineID = %d\n", goroutineID)
}
