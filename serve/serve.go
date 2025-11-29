package serve

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/zackb/updog/env"
)

type Server struct {
	*http.Server
	CertFile string
	KeyFile  string
}

func (s *Server) ListenAndServe() error {
	if s.CertFile != "" && s.KeyFile != "" {
		return s.Server.ListenAndServeTLS(s.CertFile, s.KeyFile)
	}
	return s.Server.ListenAndServe()
}

func NewHTTPServer(adder func(*http.ServeMux)) *Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK\n")
	})

	adder(handler)

	return &Server{
		Server:   &http.Server{Addr: ":" + strconv.Itoa(env.GetHTTPPort()), Handler: handler},
		CertFile: env.GetTLSCert(),
		KeyFile:  env.GetTLSKey(),
	}
}
