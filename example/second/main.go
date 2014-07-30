package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/redis"
)

func main() {

	// create our service
	t := redis.NewDirect("127.0.0.1:6379")
	s := qp.NewService("second", "one", qp.JSON, t)
	s.Handler = qp.RequestHandlerFunc(func(r *qp.Request) {
		d, _ := json.Marshal(r)
		fmt.Println("Hello from second!", string(d))
		r.Data.(map[string]interface{})["messages"] = append(r.Data.(map[string]interface{})["messages"].([]interface{}), "Hello from the second service at "+time.Now().String())

	})

	err := t.Start()
	if err != nil {
		fmt.Println("error starting transport", err)
	}
	fmt.Println("Second service started!")
	wait := make(chan struct{})

	// Set up the interrupt catch
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for _ = range c {
			signal.Stop(c)
			close(c)
			close(wait)
		}
	}()

	<-wait

	t.Stop(0)

	fmt.Println("Second service terminated!")
}
