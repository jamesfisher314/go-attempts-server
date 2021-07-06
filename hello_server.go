// Taken initially from https://www.callicoder.com/docker-golang-image-container-example/
// Because of course free software uses free examples
package main

import (
	"context"
	"errors"
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

	source, err := getSource(r)
	if check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Sources are %s\n", source)

	credPath := "/store/creds/"
	users, err := os.ReadDir(credPath)
	if check(err) {
		log.Println(err.Error())
		log.Println("FATAL: could not read the user uniquifiers")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	matchedUser := false
	for _, user := range users {
		if user.Name() == name {
			matchedUser = true
			if credbytes, err := ioutil.ReadFile(credPath + name); err != nil {
				if check(err) {
					log.Println(err.Error())
					log.Println("FATAL: could not compare the user uniquifier")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				if string(credbytes) == uniquifier {
					if err := storeIPAddress(r, w, name, uniquifier); err != nil {
						check(err)
						log.Println(err.Error())
						w.WriteHeader(http.StatusInternalServerError)
					}
					return
				} else {
					log.Printf("INFO: User provided bad uniquifier")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}
		}
	}

	if !matchedUser {
		if err := storeUniquifier(r, w, name, uniquifier); err != nil {
			if check(err) {
				log.Printf("ERROR: Failed to store user's (%s) uniquifier", name)
				log.Println(err.Error())
				return
			}
		}
		if err := storeIPAddress(r, w, name, uniquifier); err != nil {
			if check(err) {
				log.Printf("ERROR: Failed to store user's (%s) IP address", name)
				log.Println(err.Error())
				return
			}
		}
	}

	// if cred, err := ioutil.ReadFile("/go/src/static-auth/" + name); err != nil {
	// 	if check(err) {
	// 		log.Printf("Uniquifier did not match %d", len(uniquifier))
	// 		return
	// 	}
	// } else {
	// 	log.Printf("Authorization: %t", uniquifier == string(cred))
	// 	if uniquifier != string(cred) {
	// 		error_string := "401: Not authorized\n"
	// 		log.Print(error_string)
	// 		w.WriteHeader(http.StatusUnauthorized)
	// 		w.Write([]byte(error_string))
	// 		return
	// 	} else {
	// 		err := ioutil.WriteFile("/store/auth/"+name, []byte(source), 0600)
	// 		if check(err) {
	// 			return
	// 		}
	// 	}
	// }

}

func storeIPAddress(r *http.Request, w http.ResponseWriter, name string, uniquifier string) error {
	authPath := "/store/authz/"
	filePath := authPath + name

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		bytes = []byte{}
	}

	source, err := getSource(r)
	if check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		return err
	}

	contents := string(bytes)
	addresses := strings.Split(contents, "\n")

	address_map := map[string]int{}
	for _, address := range addresses {
		if len(address) > 6 {
			address_map[address] = 1
		}
	}
	address_map[source] = 1

	contents = ""
	for address, _ := range address_map {
		contents = contents + address + "\n"
	}

	err = ioutil.WriteFile(filePath, []byte(contents), 0600)
	if check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		return err
	}

	return nil
}

func storeUniquifier(r *http.Request, w http.ResponseWriter, name string, uniquifier string) error {
	credPath := "/store/creds/"
	filePath := credPath + name

	err := ioutil.WriteFile(filePath, []byte(uniquifier), 0600)
	if check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		return err
	}
	return nil
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
	os.Mkdir("/store/authz", 0777)
	os.Mkdir("/store/creds", 0777)

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

	if check(err) {
		log.Panicln(err)
		return "FATAL: Could not print this request"
	}
	return string(requestDump)
}

func check(err error) bool {
	if err != nil {
		log.Print(fmt.Println(err.Error()))
		return true
	}
	return false
}

func getSource(r *http.Request) (string, error) {
	banned := "127.0.0.1"
	remote := r.RemoteAddr
	remoteIP := strings.Split(remote, ":")
	if len(remoteIP) > 0 {
		remote = remoteIP[0]
	}
	if remote != banned {
		if len(remote) > 6 {
			return remote, nil
		}
	}

	forward := r.Header.Get("x-forwarded-for")

	if forward != banned {
		return forward, nil
	}

	return "", errors.New("FATAL: No source included in the request. How?")
}
