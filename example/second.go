package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/codecs"
	"github.com/qp/go/messages"
	"github.com/qp/go/transports"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	// create our messenger
	m := qp.NewMessenger("second", &codecs.JSON{}, transports.NewRedis("127.0.0.1:6379"))
	m.OnRequest = func(message *messages.Message) interface{} {
		fmt.Println("Second received:", message)
		message.Data.(map[string]interface{})["messages"] = append(message.Data.(map[string]interface{})["messages"].([]interface{}), "Hello from the second service at "+time.Now().String())
		wg.Done()
		return nil
	}
	m.Start()
	fmt.Println("Second service started!")
	wg.Wait()
	time.Sleep(1 * time.Second)
	m.Stop()
	fmt.Println("Second service terminated!")
}
