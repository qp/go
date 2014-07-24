package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/qp/go/codecs"
	"github.com/qp/go/transports/request"

	"github.com/qp/go"

	"github.com/stretchr/graceful"
)

func main() {

	// create our messenger
	m := qp.MakeRequestMessenger("webserver", "one", codecs.MakeJSON(), request.MakeRedis("127.0.0.1:6379"))
	err := m.Start()
	if err != nil {
		fmt.Println("error!", err)
	}

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
		msg := r.Response()
		json.NewEncoder(w).Encode(msg)
	})

	fmt.Println("Server started. Visit localhost:3001")

	graceful.Run(":3001", 10*time.Second, mux)
	m.Stop()

	fmt.Println("Server terminated.")
}
