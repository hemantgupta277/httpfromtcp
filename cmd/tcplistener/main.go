package main

import (
	"fmt"
	"log"
	"net"

	"boot.thehemantgupta.tv/httpfromtcp/internal/request"
)

// func getLinesChannel(f io.ReadCloser) <-chan string {
// 	out := make(chan string, 1)
// 	go func() {
// 		defer f.Close()
// 		defer close(out)

// 		str := ""
// 		for {
// 			data := make([]byte, 8)
// 			n, err := f.Read(data)
// 			if err != nil {
// 				break
// 			}
// 			data = data[:n]
// 			if i := bytes.IndexByte(data, '\n'); i != -1 {
// 				str += string(data[:i])
// 				data = data[i+1:]
// 				out <- str
// 				str = ""
// 			}
// 			str += string(data)
// 		}

// 		if len(str) != 0 {
// 			out <- str
// 		}
// 	}()
// 	return out
// }

// func getFileMain() {
// 	f, err := os.Open("message.txt")
// 	if err != nil {
// 		log.Fatal("error", "error", err)
// 	}

// 	lines := getLinesChannel(f)
// 	for line := range lines {
// 		fmt.Printf("read: %s\n", line)
// 	}
// }

// func main() {
// 	listener, err := net.Listen("tcp", ":42069")
// 	if err != nil {
// 		log.Fatal("error", "error", err)
// 	}

// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			log.Fatal("error", "error", err)
// 		}

// 		for line := range getLinesChannel(conn) {
// 			fmt.Printf("read: %s\n", line)
// 		}
// 	}
// }

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", "error", err)
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
		fmt.Printf("Headers:\n")
		r.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s : %s\n", n, v)
		})
	}
}
