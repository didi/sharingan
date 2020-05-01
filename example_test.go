package sharingan_test

import (
	"fmt"

	"github.com/didi/sharingan"
)

func Example() {
	sharingan.SetDelegatedFromGoRoutineID(1)

	// goroutineID = 0 「without any tag，default」
	// goroutineID = 1 「with recorder or replayer」
	goroutineID := sharingan.GetCurrentGoRoutineID()

	//Output:
	//goroutineID = 0
	fmt.Printf("goroutineID = %d\n", goroutineID)
}
