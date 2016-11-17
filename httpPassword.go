// httPassword
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Generic handler that will check the url against the mux map and
// determine what handler function to call.
type myHandler struct{}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check if the server is listening
	if stoppableServer.ServerState() == Listening {
		// if the key exists in the map (ok will be true) then call the returned func from the map in our case "hello'
		s := strings.Split(r.URL.String(), "/")
		url := s[1]

		if h, ok := mux[url]; ok {
			h(w, r)
			return
		}
		io.WriteString(w, "Request not supported: "+r.URL.String())
	} else {
		// ignore request when the server is not listening
		r.Body.Close()
	}
}

/*
*   Hash Handler function will accept /hash post request to get a job number
*   and hash the given password.
*
*	Also will handle /hash/1 get request to return the hashed password for the
*   given job number
 */
func Hash(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// this is a post so create a job number and hash the password
		r.ParseForm()                  // Parses the request body
		pass := r.Form.Get("password") // pass will be "" if parameter is not set

		jobNum := hashInfo.NextJobNumber() // get the next job number
		io.WriteString(w, strconv.FormatInt(jobNum, 10))
		// thread  off request to create a hash of the password
		go hashInfo.ProcessHashRequest(pass, jobNum)
	} else {
		// return the hashed password for the given job number
		s := strings.Split(r.URL.String(), "/")
		if len(s) >= 3 {
			// check if the password is nil and return appropriate message
			if jobNum, err := strconv.ParseInt(s[2], 10, 64); err == nil {
				// check if the jobNumber exists
				if hashedPswrd, ok := hashInfo.Password(jobNum); ok {
					// check if password is still hashing
					if hashedPswrd != "" {
						io.WriteString(w, hashedPswrd)
					} else {
						io.WriteString(w, "The password for job number ["+strconv.FormatInt(jobNum, 10)+"] is still hashing try again later.")
					}
				} else {
					// the job number is not in the map
					io.WriteString(w, "The job number {"+strconv.FormatInt(jobNum, 10)+"} doesn't exist.")
				}
			} else {
				io.WriteString(w, "Unable to get encoded password because the Job number from the request is not a valid number: "+s[2])
			}
		} else {
			io.WriteString(w, "Unable to get encoded password because the Job number is missing from the request.")
		}
	}

}

/*
*   ShowStats Handler function will accept /stats get request and return
*   and hash stats number average hash time.
 */
func ShowStats(w http.ResponseWriter, r *http.Request) {
	stats := hashInfo.requestStats.Stats()

	// switch to json data
	if statsJson, err := json.Marshal(stats); err == nil {
		io.WriteString(w, string(statsJson))
	} else {
		io.WriteString(w, "Unable to get stats: "+err.Error())
	}
}

/*
*   Closeme Handler function will accept /close get request a
*   will initiate the shutting down of the http server
 */
func Closeme(w http.ResponseWriter, r *http.Request) {
	if stoppableServer.ServerState() == Listening {
		// only close if the server is currently listening
		go stoppableServer.Close()
		closeRequestIssued <- true
		io.WriteString(w, "Sent HTTP Server close request...")
	} else {
		io.WriteString(w, "HTTP Server close request already issued...")
	}
}

// TODO these two functions can be moved to a utility package
func Round(val float64) int64 {
	if val < 0 {
		return int64(val - 0.5)
	}
	return int64(val + 0.5)
}

func DurationToMillis(val time.Duration) int64 {
	return Round(float64(val.Seconds()) * 1000)
}

var (
	mux                map[string]func(http.ResponseWriter, *http.Request)
	stoppableServer    *HttpStoppableServer
	hashInfo           *HashInfo
	closeRequestIssued chan bool
)

func main() {
	defer func(start time.Time) {
		elapsed := time.Since(start)
		fmt.Printf("This program was running for %.6fs\n", elapsed.Seconds())
		// print out the hash job stats
		fmt.Println("stats: ", hashInfo.requestStats.Stats())
	}(time.Now())

	closeRequestIssued = make(chan bool)
	hashInfo = NewHashInfo()
	fmt.Println("Starting HTTP Server...")

	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["hash"] = Hash
	mux["close"] = Closeme
	mux["stats"] = ShowStats

	server := http.Server{
		Addr:    ":8042",
		Handler: &myHandler{},
	}

	stoppableServer = NewHtmlStoppableServer(&server)
	go stoppableServer.ListenAndServe()
	fmt.Println("HTTP Server Started")

	// wait here till user issues a close request
	<-closeRequestIssued

	// wait for password hashing to complete before existing the program
	hashInfo.wg.Wait()
	fmt.Println("HTTP Server Stopped")
}
