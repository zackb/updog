package serve

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/zackb/updog/env"
)

func NewHTTPServer(adder func(*http.ServeMux)) *http.Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK\n")
	})

	adder(handler)

	return &http.Server{Addr: ":" + strconv.Itoa(env.GetHTTPPort()), Handler: handler}
}
