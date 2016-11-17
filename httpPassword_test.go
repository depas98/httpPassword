// httpPassword_test
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestHashHandler(t *testing.T) {
	hashInfo = NewHashInfo()

	data := url.Values{}
	data.Add("password", "testpass123")

	req, err := http.NewRequest("POST", "/hash", bytes.NewBufferString(data.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(Hash)

	handler.ServeHTTP(rec, req)

	expected := "1"
	if rec.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v expected %v",
			rec.Body.String(), expected)
	}

	req, err = http.NewRequest("GET", "/hash/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond) // wait a little time so the job number gets recorded
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected = "The password for job number [1] is still hashing try again later."
	if rec.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v expected %v",
			rec.Body.String(), expected)
	}

	// wait 6 seconds for the password to be hashed
	time.Sleep(6 * time.Second)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected = "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
	if rec.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v expected %v",
			rec.Body.String(), expected)
	}

	req, err = http.NewRequest("GET", "/hash/2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected = "The job number {2} doesn't exist."
	if rec.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v expected %v",
			rec.Body.String(), expected)
	}

	req, err = http.NewRequest("GET", "/hash/X2X", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected = "Unable to get encoded password because the Job number from the request is not a valid number: X2X"
	if rec.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v expected %v",
			rec.Body.String(), expected)
	}
}

func TestShowStatsHandler(t *testing.T) {
	hashInfo = NewHashInfo()

	data := url.Values{}
	data.Add("password", "testpass123")

	req, err := http.NewRequest("POST", "/hash", bytes.NewBufferString(data.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(Hash)

	// call 6 posts in a row, this also tests concurrency
	handler.ServeHTTP(rec, req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	// wait 6 seconds for the password to be hashed
	time.Sleep(6 * time.Second)

	req, err = http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec = httptest.NewRecorder()
	handler = http.HandlerFunc(ShowStats)
	handler.ServeHTTP(rec, req)

	res := RequestStats{}
	json.Unmarshal([]byte(rec.Body.String()), &res)

	expected := 6
	if res.Total != expected {
		t.Errorf("handler returned unexpected body: got %v expected %v",
			res.Total, expected)
	}
}
