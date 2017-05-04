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
)

// TODO:: CLOSE THE SERVER
var (
	httpPort = "80"
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
	// defer lconn.Close()

	buffer := bytes.NewBuffer(nil)
	// create a reader to local connection : lconn
	reader := bufio.NewReader(io.TeeReader(lconn, buffer))
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Fatalln("Could not get request... : ", err)
	}
	// intended host... might be changed to remote proxy ip address
	host := req.Host
	if !strings.Contains(host, ":") {
		host = host + ":" + httpPort
	}
	log.Println("request Host : ", host) //" from: ", req.RemoteAddr)

	// TODO :: launch separate go routines to handle read and writing to the proxy user...

	// create a connection to the remote host
	rconn, err := net.Dial("tcp", host)
	if err != nil {
		log.Println("could not access :", host, "reason:", err)
		return
	}
	// defer rconn.Close()

	// send the request...
	_, err = io.Copy(rconn, reader)
	if err != nil {
		log.Println("Error encountered during request : ", err)
		return
	}

	// read the response from the remote connection and send to local connection
	// rconn has the read method which makes it implement the io.Reader interface
	_, err = io.Copy(lconn, rconn)
	if err != nil {
		log.Println("Error encountered during request : ", err)
		return
	}
}
