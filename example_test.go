package sharingan_test

import (
	"fmt"

	"github.com/didi/sharingan"
)

func Example() {
	// goroutineID = 0 「without any tag，ignore」
	// goroutineID = 1 「with recorder or replayer」
	sharingan.SetDelegatedFromGoRoutineID(1)
	goroutineID := sharingan.GetCurrentGoRoutineID()

	//Output:
	//goroutineID = 0
	fmt.Printf("goroutineID = %d\n", goroutineID)
}
