package helper

import "os"

var isReplaying bool

// BeginReplay 开始回放
func BeginReplay() {
	os.Remove("/tmp/proxy_map")
	isReplaying = true
}

// EndReplay 结束回放
func EndReplay() {
	isReplaying = false
}

// IsReplaying 是否在回放
func IsReplaying() bool {
	return isReplaying
}
