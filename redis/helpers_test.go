package redis_test

import (
	"os/exec"
	"testing"
)

func ensureRedis(t *testing.T) {
	// is redis installed?
	err := exec.Command("which", "redis-cli").Run()
	if err != nil {
		t.Skip("skipping because redis is not installed")
	}
	// is redis running?
	err = exec.Command("redis-cli", "ping").Run()
	if err != nil {
		t.Skip("skipping because redis is not running")
	}
}
