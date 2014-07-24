package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/codecs"
	"github.com/qp/go/exchange"
	"github.com/qp/go/transports/request"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	// create our messenger
	m := qp.MakeRequestMessenger("third", "one", codecs.MakeJSON(), request.MakeRedis("127.0.0.1:6379"))
	m.OnRequest(func(channel string, request *exchange.Request) {
		d, _ := json.Marshal(request)
		fmt.Println("Hello from third!", string(d))
		request.Data.(map[string]interface{})["messages"] = append(request.Data.(map[string]interface{})["messages"].([]interface{}), "Hello from the third service at "+time.Now().String())
		wg.Done()
	}, "third")

	err := m.Start()
	if err != nil {
		fmt.Println("error!", err)
	}

	fmt.Println("Third service started!")
	wg.Wait()
	time.Sleep(1 * time.Second)
	m.Stop()
	fmt.Println("Third service terminated!")
}
