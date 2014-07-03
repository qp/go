package transports

import (
	"os/exec"
	"testing"
)

// ensure RabbitMQ conforms to Transport interface
var _ Transport = (*RabbitMQ)(nil)

// initRabbitMQ does the following:
// 1. Checks if rabbitmq is installed
// 2. Checks if rabbitmq is running
// 3. Starts it if it is not
func initRabbitMQ() bool {
	err := exec.Command("which", "rabbitmqctl").Run()
	if err != nil {
		return false
	}
	err = exec.Command("rabbitmqctl", "status").Run()
	if err != nil {
		// rabbitmq is not running. Run it.
		err = exec.Command("rabbitmq-server", "-detached").Run()
		if err != nil {
			return false
		}
	}
	return true
}

func stopRabbitMQ() {
	exec.Command("rabbitmqctl", "stop").Run()
}

func TestRabbitMQ(t *testing.T) {

}
