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
	"github.com/stretchr/slog"
)

func main() {

	// create our service
	t := redis.NewDirect("127.0.0.1:6379")

	// setup logger to Stdout
	t.SetLogger(slog.New("second", slog.Everything))

	err := qp.Service("second", "one", qp.JSON, t,
		qp.TransactionHandlerFunc(func(r *qp.Transaction) *qp.Transaction {
			d, _ := json.Marshal(r)
			fmt.Println("Hello from second!", string(d))
			r.Data.(map[string]interface{})["messages"] = append(r.Data.(map[string]interface{})["messages"].([]interface{}), "Hello from the second service at "+time.Now().String())
			return r
		}),
	)

	if err != nil {
		fmt.Println("error registering service", err)
		return
	}

	err = t.Start()
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
