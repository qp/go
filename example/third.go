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
	m := qp.NewMessenger("third", &codecs.JSON{}, transports.NewRedis("127.0.0.1:6379"))
	m.OnRequest = func(message *messages.Message) interface{} {
		fmt.Println("Third received:", message)
		message.Data.(map[string]interface{})["messages"] = append(message.Data.(map[string]interface{})["messages"].([]interface{}), "Hello from the third service at "+time.Now().String())
		wg.Done()
		return nil
	}
	m.Start()
	fmt.Println("Third service started!")
	wg.Wait()
	time.Sleep(1 * time.Second)
	m.Stop()
	fmt.Println("Third service terminated!")
}
