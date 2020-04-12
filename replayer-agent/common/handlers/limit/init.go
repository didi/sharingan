package limit

var Handler chan int

func Init() {
	Handler = make(chan int, 1)
	Handler <- 0
}
