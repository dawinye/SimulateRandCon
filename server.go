package main

import (
	"bytes"
	"fmt"
	"encoding/gob"
	"net"

	"io"
	"os"
	"strconv"
)
var m = make(map[int]net.Conn)
var vals = make(map[int][]float64)

type data struct {
	val float64
	round int
}

func handleConn(conn net.Conn) {
	tmp := make([]byte, 2)
	//fmt.Println(tmp)
	for {
		_, err := conn.Read(tmp)
		d := new(data)
		tmpBuff := bytes.NewBuffer(tmp)
		gobObj := gob.NewDecoder(tmpBuff)
		gobObj.Decode(d)
		if err == io.EOF {
			conn.Close()

			fmt.Println("One of the nodes is now faulty")

			return
		}
		fmt.Println(tmp)
		vals[d.round] = append(vals[d.round], d.val)
		fmt.Println(vals)
	}
}
func stopConsensus() {
	stopClients()
	os.Exit(10)
}
func stopClients() {
	for _, conn := range m{

		conn.Write([]byte("EXIT"))
	}
}
func main() {
	args := os.Args

	//error checking to see if the port number is provided
	if len(args) != 2 {
		fmt.Println("Please rerun the program using \"go run server.go (port number)\"")
		return
	}

	//check if the port number supplied falls within the preallocated ports
	port, err := strconv.Atoi(args[1])
	if err != nil || port < 1 || port > 65535 {
		fmt.Println("Please rerun the program using \"go run server.go (port number between 1 and 65535, inclusive)\"")
		return
	}
	l, err := net.Listen("tcp4", ":"+args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	//go stopConsensus()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		//run each client connection in a goroutine to allow concurrency
		go handleConn(conn)
	}
}
