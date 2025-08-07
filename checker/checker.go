package checker

import (
	"fmt"
	"net"
	"time"
)

// CheckPortUsingTCP tries to connect using TCP (net.DialTimeout instead of nc)
func CheckPortUsingTCP(host string, port int, timeout time.Duration) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

// CheckService checks the port 3 times before returning false, with a retry delay
func CheckService(host string, port int, retries int, retryDelay, timeout time.Duration) bool {
	for i := 0; i < retries; i++ {
		ok := CheckPortUsingTCP(host, port, timeout)
		if ok {
			return true
		}
		if i < retries-1 {
			time.Sleep(retryDelay)
		}
	}
	return false
}
