package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/qp/go/transports"

	"github.com/qp/go/codecs"

	"github.com/qp/go"

	"github.com/stretchr/graceful"
)

func main() {

	// create our messenger
	m := qp.NewMessenger("webserver", &codecs.JSON{}, transports.NewRedis("127.0.0.1:6379"))
	m.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.String() == "/favicon.ico" {
			return
		}
		data := map[string]interface{}{
			"messages": []string{"Hello from the webserver at " + time.Now().String()},
		}
		r, err := m.Request(data, "first", "second", "third")
		if err != nil {
			fmt.Fprintf(w, "Unable to make request: %v\n", err)
			return
		}
		msg := r.Message()
		if msg.HasError() {
			fmt.Fprintf(w, "An error was encountered while processing the request:\n%v\n", msg)
			return
		}
		fmt.Fprintf(w, "Request processed sucessfully:\n%v\n", msg)
	})

	fmt.Println("Server started. Visit localhost:3001")

	graceful.Run(":3001", 10*time.Second, mux)
	m.Stop()

	fmt.Println("Server terminated.")
}
