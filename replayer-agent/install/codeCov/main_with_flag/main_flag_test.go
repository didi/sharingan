// Only for the flag requirements in business server!
package main

import (
	"flag"
	"testing"
)

var systemTest *bool

func init() {
	systemTest = flag.Bool("systemTest", false, "Set to true when running system tests")
}

// Test started when the test binary is started. Only calls main.
func Test_main(t *testing.T) {
	if *systemTest {
		main()
	}
}
