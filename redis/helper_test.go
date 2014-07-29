package redis_test

import (
	"os/exec"
	"testing"
)

func ensureRedis(t *testing.T) {
	err := exec.Command("which", "redis-cli").Run()
	if err != nil {
		t.Skip("Redis not installed")
	}
	err = exec.Command("redis-cli", "ping").Run()
	if err != nil {
		t.Skip("Redis not running")
	}
}
