package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/qp/go"
	"github.com/qp/go/redis"
	"github.com/stretchr/graceful"
)

func main() {

	// create our requester
	t := redis.NewDirect("127.0.0.1:6379")
	r := qp.NewRequester("webserver", "one", qp.JSON, t)
	t.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.String() == "/favicon.ico" {
			return
		}
		obj := map[string]interface{}{
			"messages": []string{"Hello from the webserver at " + time.Now().String()},
		}
		f, err := r.Issue([]string{"first", "second", "third"}, obj)
		if err != nil {
			fmt.Fprintf(w, "error issuing request: %v\n", err)
			return
		}

		msg, err := f.Response(1 * time.Second)
		if err != nil {
			fmt.Fprintf(w, "error getting response: %v\n", err)
			return
		}
		json.NewEncoder(w).Encode(msg)
	})

	fmt.Println("Server started. Visit localhost:3001")

	graceful.Run(":3001", 10*time.Second, mux)
	t.Stop(0)

	fmt.Println("Server terminated.")
}
