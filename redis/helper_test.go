package redis_test

import (
	"os/exec"
	"time"
)

// initRedis does the following:
// 1. Checks if Redis is installed
// 2. Checks if Redis is running
// 3. Starts it if it is not
func initRedis() bool {
	err := exec.Command("which", "redis-cli").Run()
	if err != nil {
		return false
	}
	err = exec.Command("redis-cli", "ping").Run()
	if err != nil {
		// Redis is not running. Run it.
		err = exec.Command("redis-server", "--daemonize", "yes").Run()
		if err != nil {
			return false
		}
		time.Sleep(200 * time.Millisecond)
	}
	return true
}

func stopRedis() {
	exec.Command("redis-cli", "shutdown").Run()
}
