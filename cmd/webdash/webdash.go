package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/hsson/cluster-monitor/pkg/clusterinfo"
)

const (
	writeTimeout = 15 * time.Second
	readTimeout  = 15 * time.Second
	waitTimeout  = 15 * time.Second
	idleTimeout  = 60 * time.Second

	defaultPort = "8000"
)

var _indexTemplate = template.Must(template.ParseFiles("index.html"))
var _clusterClient clusterinfo.Client

type pageData struct {
	Nodes []clusterinfo.Node
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/healthz", health).Methods(http.MethodGet)
	r.HandleFunc("/", index).Methods(http.MethodGet)
	r.NotFoundHandler = http.HandlerFunc(notFound)
	r.Use(requestLoggingMiddleware)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + getPort(),
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Println("starting server")

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("error: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()

	srv.Shutdown(ctx)
	log.Println("shutting down server")
	os.Exit(0)
}

func index(w http.ResponseWriter, r *http.Request) {
	nodes, err := _clusterClient.Nodes().List()
	if err != nil {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Internal server error — failed to get nodes: %v", err)
		return
	}
	err = _indexTemplate.Execute(w, pageData{
		Nodes: nodes.All(),
	})
	if err != nil {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Internal server error — failed to render page: %v", err)
	} else {
		w.Header().Set("Content-Type", "text/html")
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func notFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("404 not found: %s %s", r.Method, r.RequestURI)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(404)
	fmt.Fprint(w, "404 not found")
}

func requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func getPort() string {
	const portKey = "CONN_PORT"
	if p, found := os.LookupEnv(portKey); found {
		return p
	}
	return defaultPort
}
