package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// https://www.aozora.gr.jp/cards/000042/files/2444_10269.html
var contents = []string{
	"どこかへ旅行がしてみたくなる。しかし別にどこというきまったあてがない。",
	"そういう時に旅行案内記の類をあけて見ると、あるいは海浜、あるいは山間の湖水、あるいは温泉といったように、行くべき所がさまざま有りすぎるほどある。",
	"そこでまずかりに温泉なら温泉ときめて、温泉の部を少し詳しく見て行くと、各温泉の水質や効能、周囲の形勝名所旧跡などのだいたいがざっとわかる。",
	"しかしもう少し詳しく具体的の事が知りたくなって、今度は温泉専門の案内書を捜し出して読んでみる。",
	"そうするとまずぼんやりとおおよその見当がついて来るが、いくら詳細な案内記を丁寧に読んでみたところで、結局ほんとうのところは自分で行って見なければわかるはずはない。",
	"もしもそれがわかるようならば、うちで書物だけ読んでいればわざわざ出かける必要はないと言ってもいい。",
	"次には念のためにいろいろの人の話を聞いてみても、人によってかなり言う事がちがっていて、だれのオーソリティを信じていいかわからなくなってしまう。",
	"それでさんざんに調べた最後には、つまりいいかげんに、賽さいでも投げると同じような偶然な機縁によって目的の地をどうにかきめるほかはない。",
}

func isGZipAcceptable(request *http.Request) bool {
	return strings.Index(strings.Join(request.Header["Accept-Encoding"], ","), "gzip") != -1
}

func processSessionWithChunk(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Accept %v\n", conn.RemoteAddr())

	for {
		request, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		dump, err := httputil.DumpRequest(request, true)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(dump))

		fmt.Fprintf(conn, strings.Join([]string{
			"HTTP/1.1 200 OK",
			"Content-Type: text/plain",
			"Transfer-Encoding: chunked",
			"",
			"",
		}, "\r\n"))
		for _, content := range contents {
			bytes := []byte(content)
			fmt.Fprintf(conn, "%x\r\n%s\r\n", len(bytes), content)
		}
		fmt.Fprintf(conn, "0\r\n\r\n")
	}
}

func processSession(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Accept %v¥n", conn.RemoteAddr())

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		request, err := http.ReadRequest(bufio.NewReader(conn))
		if err != nil {
			neterr, ok := err.(net.Error)
			if ok && neterr.Timeout() {
				fmt.Println("timeout")
				break
			} else if err == io.EOF {
				break
			}
			panic(err)
		}

		dump, err := httputil.DumpRequest(request, true)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(dump))

		response := http.Response{
			StatusCode: 200,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
		}

		if isGZipAcceptable(request) {
			content := "Hello World(gzipped)\n"
			var buffer bytes.Buffer
			writer := gzip.NewWriter(&buffer)
			io.WriteString(writer, content)
			writer.Close()
			response.Body = ioutil.NopCloser(&buffer)
			response.ContentLength = int64(buffer.Len())
			response.Header.Set("Content-Encoding", "gzip")
		} else {
			content := "Hello World\n"
			response.Body = ioutil.NopCloser(strings.NewReader(content))
			response.ContentLength = int64(len(content))
		}
		response.Write(conn)
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8888")
	if err != nil {
		panic(err)
	}

	fmt.Println("Server is running at localhost:8888")

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		// go processSession(conn)
		go processSessionWithChunk(conn)
	}
}
