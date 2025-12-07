package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"boot.thehemantgupta.tv/httpfromtcp/internal/headers"
	"boot.thehemantgupta.tv/httpfromtcp/internal/request"
	"boot.thehemantgupta.tv/httpfromtcp/internal/response"
	"boot.thehemantgupta.tv/httpfromtcp/internal/server"
)

func toStr(bytes []byte) string {
	out := ""
	for _, b := range bytes {
		out += fmt.Sprintf("%02x", b)
	}
	return out
}

const port = 42069

func respond400() []byte {
	return []byte(`
		<html>
			<head>
				<title>400 Bad Request</title>
			</head>
			<body>
				<h1>Bad Request</h1>
				<p>Your request honestly kinda sucked.</p>
			</body>
		</html>
	`)
}

func respond500() []byte {
	return []byte(`
		<html>
			<head>
				<title>500 Internal Server Error</title>
			</head>
			<body>
				<h1>Internal Server Error</h1>
				<p>Okay, you know what? This one is on me.</p>
			</body>
		</html>
	`)
}

func respond200() []byte {
	return []byte(`
		<html>
			<head>
				<title>200 OK</title>
			</head>
			<body>
				<h1>Success!</h1>
				<p>Your request was an absolute banger.</p>
			</body>
		</html>
	`)
}

// var body []byte = nil
// 	var status response.StatusCode = response.StatusOK
// 	if errHandler != nil {
// 		status = errHandler.StatusCode
// 		body = []byte(errHandler.Message)
// 	} else {
// 		body = writer.Bytes()
// 	}
// 	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))

// 	response.WriteStatusLine(conn, status)
// 	response.WriteHeaders(conn, headers)
// 	conn.Write(body)

func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)
		body := respond200()
		status := response.StatusOK
		if req.RequestLine.RequestTarget == "/yourproblem" {
			body = respond400()
			status = response.StatusBadRequest
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			body = respond500()
			status = response.StatusInternalServerError
		} else if req.RequestLine.RequestTarget == "/video" {
			f, _ := os.ReadFile("assets/vim.mp4")
			h.Replace("content-type", "video/mp4")
			h.Replace("content-length", fmt.Sprintf("%d", len(f)))
			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(*h)
			w.WriteBody(f)
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
			target := req.RequestLine.RequestTarget
			res, err := http.Get("https://httpbin.org/" + target[len("/httpbin/"):])
			if err != nil {
				body = respond500()
				status = response.StatusInternalServerError
			} else {
				w.WriteStatusLine(response.StatusOK)
				h.Delete("content-length")
				h.Set("transfer-encoding", "chunked")
				h.Replace("content-type", "text/plain")
				w.WriteHeaders(*h)

				fullBody := []byte{}
				for {
					data := make([]byte, 32)
					n, err := res.Body.Read(data)
					if err != nil {
						break
					}
					fullBody = append(fullBody, data[:n]...)
					w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
					w.WriteBody(data[:n])
					w.WriteBody([]byte("\r\n"))
				}
				w.WriteBody([]byte("0\r\n"))
				trailers := headers.NewHeaders()
				sha := sha256.Sum256(fullBody)
				trailers.Set("X-Content-SHA256", toStr(sha[:]))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
				w.WriteHeaders(*trailers)
				w.WriteBody([]byte("\r\n"))
				return
			}
		}
		w.WriteStatusLine(status)
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		h.Replace("Content-Type", "text/html")
		w.WriteHeaders(*h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
