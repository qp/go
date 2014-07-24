package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/redis"
)

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	// create our messenger
	m := qp.NewRequester("first", "one", qp.JSON, redis.NewReqTransport("127.0.0.1:6379"))
	m.OnRequest(func(channel string, request *qp.Request) {
		d, _ := json.Marshal(request)
		fmt.Println("Hello from first!", string(d))
		request.Data.(map[string]interface{})["messages"] = append(request.Data.(map[string]interface{})["messages"].([]interface{}), "Hello from the first service at "+time.Now().String())
		wg.Done()
	}, "first")

	err := m.Start()
	if err != nil {
		fmt.Println("error!", err)
	}
	fmt.Println("First service started!")
	wg.Wait()
	time.Sleep(1 * time.Second)
	m.Stop()
	fmt.Println("First service terminated!")
}
