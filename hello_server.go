// Taken initially from https://www.callicoder.com/docker-golang-image-container-example/
// Because of course free software uses free examples
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"
)

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")
	if name == "" {
		name = "Guest"
	}
	log.Printf("Received request for %s\n", name)
	w.Write([]byte(fmt.Sprintf("Hello, %s\n", name)))
}

func registrar(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")
	uniquifier := query.Get("uniquifier")
	log.Printf("Received request to register name '%s' with uniquifier length %d:\n%s%s",
		name,
		len(uniquifier),
		printRequest(r, false),
		"remote-addr: "+r.RemoteAddr)
	failed := false
	if name == "" {
		error_string := "Failure: Must include 'name' query\n"
		log.Print(error_string)
		w.Write([]byte(error_string))
		failed = true
	}
	if uniquifier == "" || len(uniquifier) < 16 {
		error_string := fmt.Sprintf("Failure: Uniquifier is too short at length %d\n", len(uniquifier))
		log.Print(error_string)
		w.Write([]byte(error_string))
		failed = true
	}
	if failed {
		return
	}

	source := getSource(r)

	if cred, err := ioutil.ReadFile("/go/src/static-auth/" + name); err != nil {
		check(err)
	} else {
		if uniquifier != string(cred) {
			error_string := "401: Not authorized\n"
			log.Print(error_string)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(error_string))
			return
		} else {
			err := ioutil.WriteFile("/store/auth/"+name, []byte(source), 0600)
			check(err)
		}
	}

}

func authenticator(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")
	uniquifier := query.Get("uniquifier")
	log.Printf("Received request to confirm registration with name '%s' and uniquifier length '%d'\n", name, len(uniquifier))
}

func main() {
	// Create Server and Route Handlers
	r := mux.NewRouter()

	r.HandleFunc("/", handler)
	r.HandleFunc("/register", registrar)
	r.HandleFunc("/check", authenticator)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Configure Logging
	LOG_FILE_ROOT := os.Getenv("LOG_FILE_ROOT")
	if LOG_FILE_ROOT != "" {
		os.MkdirAll(LOG_FILE_ROOT, 0777)
	}
	LOG_FILE_LOCATION := os.Getenv("LOG_FILE_LOCATION")
	if LOG_FILE_LOCATION != "" {
		ioutil.WriteFile(LOG_FILE_LOCATION, []byte(""), 0777)
		log.SetOutput(&lumberjack.Logger{
			Filename:   LOG_FILE_LOCATION,
			MaxSize:    50, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		})
	}

	// Configure registrations
	os.Mkdir("/store/auth", 0700)

	// Start Server
	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}

func printRequest(r *http.Request, hasContent bool) string {
	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, hasContent)
	check(err)
	return string(requestDump)
}

func check(err error) {
	if err != nil {
		log.Print(fmt.Println(err))
	}
}

func getSource(r *http.Request) string {
	banned := "127.0.0.1"
	remote := r.RemoteAddr
	remoteIP := strings.Split(remote, ":")
	forward := r.Header.Get("x-forwarded-for")
	log.Printf("%s%s%s%s", banned, remote, remoteIP, forward)
	return "nothing"
}
