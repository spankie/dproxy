package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// TODO:: CLOSE THE SERVER

// TODO :: USE `context` FOR CONTROL
// BUT FIRST LETS USE NORMAL OUR OWN CHANNELS FOR LEARNING SAKE

const (
	httpPort = "80"
	// set the timeout ...
	timeout = time.Second * 60
	//isConn = make(chan int)
)

func init() {
	log.SetPrefix("dproxy: ")
}

func main() {
	// set up the port for the proxy to listen on...
	localPort := "8081"

	// indicating starting of the proxy...
	log.Println("dproxy listening on :", localPort)

	// create a tcp listener, listening for a connection
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", localPort))
	if err != nil {
		// end the program if it can not listen...
		log.Fatalln("Could not create a listener for port ", localPort)
	}

	// start a go routine to control each proxy request

	for {
		// accept each incomming connection...
		conn, err := l.Accept()
		if err != nil {
			// the connection could not be accepted...
			log.Println("Could not create a connection...")
			// skip the loop and wait for another request...
			// I do not want to end the program because of this...
			continue
		}
		// route the proxy to the remote address and back...
		go proxy(conn)
	}
}

func proxy(lconn net.Conn) {
	connDone := make(chan int)

	defer lconn.Close()

	buffer := bytes.NewBuffer(nil)
	// create a reader to local connection : lconn
	reader := bufio.NewReader(io.TeeReader(lconn, buffer))
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println("Could not get request... : ", err)
		return
	}
	// intended host... might be changed to remote proxy ip address
	host := req.Host
	if !strings.Contains(host, ":") {
		host = host + ":" + httpPort
	}
	log.Println("request Host : ", host) //" from: ", req.RemoteAddr)

	// TODO :: launch separate go routines to handle read and writing to the proxy user...

	// create a connection to the remote host
	rconn, err := net.DialTimeout("tcp", "google.com:80", timeout)
	if err != nil {
		log.Println("could not access :", host, "reason:", err)
		return
	}

	go func() {
		defer func() { connDone <- 1 }()
		for {
			lconn.SetReadDeadline(time.Now().Add(timeout))
			rconn.SetWriteDeadline(time.Now().Add(timeout))
			// send the request using 8KiB for each copy...
			w, err := io.CopyN(rconn, lconn, 8*1<<10)
			log.Println("sent:", w)
			if isNormalTerminationError(err) {
				log.Println("sent: copy err : ", err)
				// connDone <- 1
				return
			}
			if err != nil {
				log.Println("Error encountered during request : ", err)

				return
			}
		}
	}()
	// kama kawaida : kagwe
	// read the response from the remote connection and send to local connection
	// rconn has the read method which makes it implement the io.Reader interface
	go func() {
		defer func() { connDone <- 1 }()
		for {
			rconn.SetReadDeadline(time.Now().Add(timeout))
			lconn.SetWriteDeadline(time.Now().Add(timeout))
			// retrieve response 8Kib at a time ....
			w, err := io.CopyN(lconn, rconn, 8*1<<10)
			log.Println("received:", w)
			if isNormalTerminationError(err) {
				log.Println("receive: copy err : ", err)
				connDone <- 1
				return
			}
			if err != nil {
				log.Println("Error encountered during request : ", err)
				connDone <- 1
				return
			}
		}
	}()

	// semaphore to control the read and write go routines
	for i := 1; i <= 2; i++ {
		<-connDone
	}
	rconn.Close()
	log.Println("proxy number # ended")
}

// function from https://github.com/aybabtme/portproxy : portproxy.go
func isNormalTerminationError(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	e, ok := err.(*net.OpError)
	if ok && e.Timeout() {
		return true
	}

	for _, cause := range []string{
		"use of closed network connection",
		"broken pipe",
		"connection reset by peer",
	} {
		if strings.Contains(err.Error(), cause) {
			return true
		}
	}

	return false
}
