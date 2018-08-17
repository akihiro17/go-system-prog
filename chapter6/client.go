package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

func main() {
	sendMessages := []string{
		"msg1",
		"msg2",
		"msg3",
	}
	current := 0
	var conn net.Conn = nil
	for {
		var err error
		if conn == nil {
			conn, err = net.Dial("tcp", "localhost:8888")
			if err != nil {
				panic(err)
			}
			fmt.Printf("Access %d\n", current)
		}
		request, err :=
			http.NewRequest("POST", "htttp://localhost:8888", strings.NewReader(sendMessages[current]))
		if err != nil {
			panic(err)
		}
		// gzip圧縮
		request.Header.Set("Accept-Encoding", "gzip")

		request.Write(conn)
		response, err := http.ReadResponse(bufio.NewReader(conn), request)
		if err != nil {
			panic(err)
		}

		dump, err := httputil.DumpResponse(response, false)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(dump))
		defer response.Body.Close()

		if response.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body)
			if err != nil {
				panic(err)
			}
			io.Copy(os.Stdout, reader)
			reader.Close()
		} else {
			io.Copy(os.Stdout, response.Body)
		}

		current++
		if current == len(sendMessages) {
			break
		}
	}
	conn.Close()
}
