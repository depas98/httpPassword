// httpPasswordService
package main

import (
	"crypto/sha512"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"
)

// HashInfo will contain the job id for next request, keeps a map of all hash request,
// and the hash request statistics (number of hash req and average time to hash password)
type HashInfo struct {
	nextJobId int64 // accessed atomically, counter that is used to hand out new job #s

	// map of job numbers to the encoded password,
	// if returns "" then the password is still being hashed
	hashJobs     map[int64]string
	hashJobsLock sync.Mutex
	wg           *sync.WaitGroup // wg used for hashing pswrd
	wgRWLock     sync.RWMutex
	requestStats *HashRequestStats
}

// NewHashInfo initiliazes the HashInfo struct
func NewHashInfo() *HashInfo {
	return &HashInfo{
		nextJobId:    0,
		hashJobs:     make(map[int64]string),
		wg:           &sync.WaitGroup{},
		requestStats: &HashRequestStats{},
	}
}

// This will get the next job number for the Hash request
func (hi *HashInfo) NextJobNumber() int64 {
	return atomic.AddInt64(&hi.nextJobId, 1)
}

// Given a password it returns the hashed password using sha512
// and the time it took to hash the password
func (hi *HashInfo) hashPassword(password string) (string, time.Duration) {
	startTime := time.Now()
	time.Sleep(5 * time.Second) // wait five seconds before hashing
	passwordToHash := []byte(password)
	hashedpassword := sha512.Sum512(passwordToHash)
	hashedpasswordStr := hex.EncodeToString(hashedpassword[:])

	elapsed := time.Since(startTime)
	return hashedpasswordStr, elapsed
}

/*
* Given a job number and a password, will first store the
* job number into the hashJobs map with a blank password
* then will call to hash the password and then will store the
* hashed password into the map for the given job number
 */
func (hi *HashInfo) ProcessHashRequest(password string, jobNumber int64) {
	hi.startHashingWg()

	// save the job number to map
	hi.hashJobsLock.Lock()
	hi.hashJobs[jobNumber] = ""
	hi.hashJobsLock.Unlock()

	hashedPswrd, duration := hi.hashPassword(password)

	//add the job num and hashed password to the jobs map
	hi.hashJobsLock.Lock()
	hi.hashJobs[jobNumber] = hashedPswrd
	hi.hashJobsLock.Unlock()

	hi.requestStats.Record(duration) // record the stats
	hi.finishHashingWg()
}

// startHashingWg increments the HashInfo's WaitGroup. Needed to support
// multi web request for password hashing
func (hi *HashInfo) startHashingWg() {
	hi.wgRWLock.Lock()
	hi.wg.Add(1)
	hi.wgRWLock.Unlock()
}

// finishHashingWg decrements the HashInfo's WaitGroup.
// Use this to complement startHashingWg().
func (hi *HashInfo) finishHashingWg() {
	hi.wgRWLock.Lock()
	hi.wg.Done()
	hi.wgRWLock.Unlock()
}

/*
*   Given a jobnumber will return the password anf a boolean.
*   If false is returned then the job number does not exist.
*   If empty string is return for password then the password
*   is still hashing.
 */
func (hi *HashInfo) Password(jobNumber int64) (string, bool) {
	hi.hashJobsLock.Lock()
	hashedpassword, ok := hi.hashJobs[jobNumber]
	hi.hashJobsLock.Unlock()
	if ok {
		return hashedpassword, true

	} else {
		return "", false
	}

}

// This struct holds the raw hashing stats data
type HashRequestStats struct {
	totalRequest int32
	totalTimeMs  int64
	statsRWLock  sync.RWMutex
}

func (hs *HashRequestStats) Record(duration time.Duration) {
	// need stats lock here
	hs.statsRWLock.Lock()

	hs.totalRequest = hs.totalRequest + 1
	hs.totalTimeMs = hs.totalTimeMs + DurationToMillis(duration)

	hs.statsRWLock.Unlock()
}

// This struct is used for marshalling the Stats json data
type RequestStats struct {
	Total   int `json:"total"`
	Average int `json:"average"`
}

// Function that will return the stats (total, avg)
func (hs *HashRequestStats) Stats() *RequestStats {
	avg := int64(0)
	if hs.totalRequest != 0 {
		// protect against divide by 0
		avg = hs.totalTimeMs / int64(hs.totalRequest)
	}

	stats := &RequestStats{
		Total:   int(hs.totalRequest),
		Average: int(avg),
	}

	return stats
}
