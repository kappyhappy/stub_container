package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	appName := getAppName()
	log.Printf("Running stub container for %s service", appName)

	r := mux.NewRouter()
	r.HandleFunc("/nettest/{host}/{port:[0-9]+}", newNetTestHandler())
	r.PathPrefix("/").HandlerFunc(newCatchAllHandler())

	addr := getListenAddr()
	if err := listenAndServe(addr, r, shouldListenHTTPS()); err != nil {
		log.Fatalf("unexpected error : %v", err)
	}
}

func listenAndServe(addr string, handler http.Handler, ssl bool) error {
	if ssl {
		return http.ListenAndServeTLS(addr, "/server.crt", "/server.key", handler)
	}

	return http.ListenAndServe(addr, handler)
}

func newCatchAllHandler() http.HandlerFunc {
	appName := getAppName()

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "stub container for %s service\n", appName)
	}
}

func newNetTestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		host := vars["host"]
		port := vars["port"]

		if !isReachable(host, port) {
			w.WriteHeader(http.StatusGatewayTimeout)
			fmt.Fprintf(w, "Failed to establish tcp connection to %s:%s\n", host, port)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully established tcp connection to %s:%s\n", host, port)
	}
}

func getListenAddr() string {
	port, found := os.LookupEnv("LISTEN_PORT")
	if !found {
		return "0.0.0.0:3000"
	}

	return "0.0.0.0:" + port
}
func getAppName() string {
	name, found := os.LookupEnv("APP_NAME")
	if !found {
		return "undefined"
	}
	return name
}

func shouldListenHTTPS() bool {
	val, found := os.LookupEnv("LISTEN_HTTPS")
	return found && val == "true"
}

func isReachable(host, port string) bool {
	timeout := time.Duration(3) * time.Second
	conn, err := net.DialTimeout("tcp", host+":"+port, timeout)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}
