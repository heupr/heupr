package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Integration is an interface for the backend resource that the API calls
type Integration interface {
	Learn(input []byte) ([]byte, error)
	Predict(input []byte) ([]byte, error)
}

func learn(i Integration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
		}

		// outline:
		// [ ] call feature insert
		// [ ] add returned feature to body

		result, err := i.Learn(body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Learn method call error: %s", err.Error()), http.StatusInternalServerError)
		} else {
			fmt.Fprint(w, string(result))
		}
	}
}

func predict(i Integration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
		}

		// outline:
		// [ ] call feature update
		// [ ] add returned feature to body

		result, err := i.Predict(body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Predict method call error: %s", err.Error()), http.StatusInternalServerError)
		} else {
			fmt.Fprint(w, string(result))
		}
	}
}

// Server hosts the RESTful API
type Server struct {
	server    http.Server
	processor processor
}

// New instantiates a new server struct without starting it
func New(i Integration, addr string) *Server {
	r := mux.NewRouter()
	r.HandleFunc("/learn", learn(i)).Methods("POST").Schemes("https")
	r.HandleFunc("/predict", predict(i)).Methods("POST").Schemes("https")

	p, err := newProcessor("heupr.bleve")
	if err != nil {
		log.Fatalf("error calling new processor: %s", err.Error())
	}

	s := &Server{
		server: http.Server{
			Addr:         addr,
			Handler:      r,
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		},
		processor: p,
	}

	return s
}

// Start triggers the server to listen and serve
func (s *Server) Start() {
	s.server.ListenAndServe()
}

// Stop closes down the server
func (s *Server) Stop() {
	s.server.Close()
}
