package helper

func CInfo(msg string) string {
	return msg
	// return "\033[;32m" + msg + "\033[0m"
}

func CWarn(msg string) string {
	return msg
	// return "\033[;33m" + msg + "\033[0m"
}

func CErr(msg string) string {
	return msg
	// return "\033[;31m" + msg + "\033[0m"
}
