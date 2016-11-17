package main

import (
	"net"
	"net/http"
	"sync"
)

// enum for the server state
type State int

const (
	NotListening State = iota - 1 // server was never started
	Listening
	Closing // no more requesst accepted, server is closing
	Closed
)

// HttpStopableServer wraps a net/http.Server
// and contains the state of the server.
//
// Initialized by calling NewHtmlServer.
type HttpStoppableServer struct {
	server       *http.Server
	serverState  State
	stop         chan bool
	stopFinished chan bool
	serverRWLock sync.RWMutex
}

// NewHtmlServer wraps an existing http.Server object and returns a
// HttpStopableServer.
func NewHtmlStoppableServer(s *http.Server) *HttpStoppableServer {
	return &HttpStoppableServer{
		server:       s,
		serverState:  NotListening, // start this at -1 (never started), and change to 0 (Listening) when they call Serve
		stop:         make(chan bool),
		stopFinished: make(chan bool, 1),
	}
}

// ListenAndServe is same as net/http.Serve.ListenAndServe.
func (s *HttpStoppableServer) ListenAndServe() error {
	addr := s.server.Addr
	if addr == "" {
		addr = ":http"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return s.Serve(listener)
}

// Wraps net/http.Server.Serve
func (s *HttpStoppableServer) Serve(listener net.Listener) error {
	// stops the listener when close request for the server is issued
	go s.stopListener(listener)

	// ConnState specifies an optional callback function that is
	// called when a client connection changes state.
	// It can have 5 states (StateNew, StateActive, StateIdle, StateHijacked, StateClosed)
	s.server.ConnState = func(conn net.Conn, connState http.ConnState) {
		if connState == http.StateActive && s.serverState != Listening {
			// For new active connections and if Server state is not listening then close the connection
			conn.Close()
		}
	}

	s.serverRWLock.Lock()
	s.serverState = Listening // set that the server is now listening
	s.serverRWLock.Unlock()
	err := s.server.Serve(listener)
	// Dont' report the error when server is not listening
	if err != nil && s.serverState != Listening {
		err = nil
	}

	s.stopFinished <- true
	return err
}

// This will stop the listener associated with the server
//  when it gets the stop signal
func (s *HttpStoppableServer) stopListener(listener net.Listener) {
	s.stop <- true // wait for the stop signal
	close(s.stop)
	s.server.SetKeepAlivesEnabled(false) // shuting down, don't have kept alive connections anymore
	listener.Close()
}

// Close stops the server from accepting new requests and begins shutting down
// It returns true if it's the first time Close is called.
func (s *HttpStoppableServer) Close() bool {
	result := <-s.stop
	s.serverRWLock.Lock()
	s.serverState = Closing // set that the server is closing
	s.serverRWLock.Unlock()
	<-s.stopFinished
	s.serverRWLock.Lock()
	s.serverState = Closed // set that the server is closed
	s.serverRWLock.Unlock()
	return result
}

// return the state of the server
func (s *HttpStoppableServer) ServerState() State {
	s.serverRWLock.RLock()
	defer s.serverRWLock.RUnlock()
	return s.serverState
}
