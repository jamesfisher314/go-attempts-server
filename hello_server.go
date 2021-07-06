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
	log.Printf("INFO: root request from %s\n", name)
	w.Write([]byte(fmt.Sprintf("Hello, %s\n", name)))
}

func registrar(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")
	uniquifier := query.Get("uniquifier")
	source, err := getSource(r)
	response := http.StatusNoContent
	if check(err) {
		response = http.StatusInternalServerError
		w.WriteHeader(response)
		log.Printf("FATAL: %d Encounterd %s while parsing sources", response, err.Error())
		return
	}

	log.Printf("INFO: register name '%s' with uniquifier length %d from source: %s",
		name,
		len(uniquifier),
		"remote-addr: "+source)
	failed := false
	error_string := ""
	if name == "" {
		error_string = "400 Bad Request: Must include 'name' query\n"
		failed = true
	}
	if uniquifier == "" || len(uniquifier) < 16 {
		error_string = fmt.Sprintf("400 Bad Request: Uniquifier is too short at length %d\n", len(uniquifier))
		failed = true
	}
	if failed {
		response = http.StatusBadRequest
		w.WriteHeader(response)
		log.Print(error_string)
		w.Write([]byte(error_string))
		return
	}

	matchedUser := false
	response, matched, authZed, ok := confirmToken(name, uniquifier, source)
	if !ok || !authZed {
		w.WriteHeader(response)
		return
	} else {
		matchedUser = matched
	}
	if !matchedUser {
		if response, err := storeUniquifier(name, uniquifier); err != nil {
			if check(err) {
				log.Printf("ERROR: Failed to store user's (%s) uniquifier", name)
				log.Println(err.Error())
				w.WriteHeader(response)
				return
			}
		}
		if response, err := storeIPAddress(name, uniquifier, source); err != nil {
			if check(err) {
				log.Printf("ERROR: Failed to store user's (%s) IP address", name)
				log.Println(err.Error())
				w.WriteHeader(response)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func confirmToken(name string, uniquifier string, source string) (int, bool, bool, bool) {
	credPath := "/store/creds/"
	users, err := os.ReadDir(credPath)
	if check(err) {
		log.Println(err.Error())
		log.Println("FATAL: could not read the user uniquifiers")
		return http.StatusInternalServerError, false, false, false // Code, matched the user, authorized, ok
	}

	matchedUser := false
	for _, user := range users {
		if user.Name() == name {
			matchedUser = true
			if credbytes, err := ioutil.ReadFile(credPath + name); err != nil {
				if check(err) {
					log.Println(err.Error())
					log.Println("FATAL: could not compare the user uniquifier")
					return http.StatusInternalServerError, matchedUser, false, false
				}
			} else { // The user's uniquifier is loaded; does it match the request's uniquifier?
				if string(credbytes) == uniquifier {
					return http.StatusNoContent, matchedUser, true, true
				} else {
					log.Printf("INFO: User provided bad uniquifier")
					return http.StatusUnauthorized, matchedUser, false, true
				}
			}
		}
	}

	return http.StatusUnauthorized, matchedUser, false, true
}

func storeIPAddress(name string, uniquifier string, source string) (int, error) {
	authPath := "/store/authz/"
	filePath := authPath + name

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		bytes = []byte{}
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
		log.Println(err.Error())
		return http.StatusInternalServerError, err
	}

	return http.StatusNoContent, nil
}

func storeUniquifier(name string, uniquifier string) (int, error) {
	credPath := "/store/creds/"
	filePath := credPath + name

	err := ioutil.WriteFile(filePath, []byte(uniquifier), 0600)
	if check(err) {
		log.Println(err.Error())
		return http.StatusInternalServerError, err
	}
	return http.StatusNoContent, nil
}

func authenticator(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")
	uniquifier := query.Get("uniquifier")
	source, err := getSource(r)
	if check(err) {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	response, matched, authZed, ok := confirmToken(name, uniquifier, source)
	log.Printf("INFO: confirm registration with name '%s' and uniquifier length '%d'; response %d matched %t authZed %t ok %t\n",
		name,
		len(uniquifier),
		response,
		matched,
		authZed,
		ok)
	w.WriteHeader(response)
}

func main() {
	// Create Server and Route Handlers
	r := mux.NewRouter()

	r.HandleFunc("/", handler)
	r.HandleFunc("/register", registrar)
	r.HandleFunc("/check", authenticator)
	// r.HandleFunc("/share/add", addShare)
	// r.HandleFunc("/share/delete", deleteShare)
	// r.HandleFunc("/share", checkShare)
	// r.HandleFunc("/subscribe/add", addSubscribe)
	// r.HandleFunc("/subscribe/delete", unSubscribe)
	// r.HandleFunc("/subscribe", checkSubscribe)

	// r.HandleFunc("/manifest/add", addManifest)
	// r.HandleFunc("/manifest/get", getManifest)
	// r.HandleFunc("/blob/add", addBlob)
	// r.HandleFunc("/blob/get", getBlob)

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
	os.Mkdir("/store/cache", 0777)

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
