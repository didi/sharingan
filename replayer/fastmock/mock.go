package fastmock

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

/* mock server信息*/

// 默认mock配置
const defaultMockIP = "127.0.0.1"
const defaultMockPort = "3515"

var (
	mockSaAddr [4]byte
	mockSaPort int
	mockIP     string
	mockPort   string
	mockAddr   string
)

func init() {
	setMockIP()
	setMockPort()

	mockAddr = mockIP + ":" + mockPort
}

// setMockIP set mockPort
func setMockIP() error {
	mIP := os.Getenv("REPLAYER_MOCK_IP")
	if mIP == "" {
		mIP = defaultMockIP
	}
	splits := strings.Split(mIP, ".")
	if len(splits) != 4 {
		return errors.New("check ip format failed")
	}

	// mockSaAddr
	for i := range splits {
		bi, err := strconv.Atoi(splits[i])
		if err != nil {
			return err
		}
		mockSaAddr[i] = byte(bi)
	}

	mockIP = mIP
	return nil
}

// setMockPort set mockPort
func setMockPort() error {
	mPort := os.Getenv("REPLAYER_MOCK_PORT")
	if mPort == "" {
		mPort = defaultMockPort
	}
	port, err := strconv.Atoi(mPort)
	if err != nil {
		return err
	}

	mockSaPort = port
	mockPort = mPort
	return nil
}
