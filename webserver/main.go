// https://codingchallenges.fyi/challenges/challenge-webserver

package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/xTaube/coding-challenges/webserver/src/file"
	"github.com/xTaube/coding-challenges/webserver/src/requests"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("You need to specify port.\n")
	}

	port := os.Args[1]
	server, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("TCP server listening on %s\n", server.Addr())

	for {
		client, err := server.Accept()
		if err != nil {
			log.Println("Could not connect with client: %", err)
			continue
		}

		go handleClient(client)
	}
}

func handleClient(client net.Conn) {
	defer client.Close()

	fmt.Printf("Handling client with addres %s\n", client.RemoteAddr())

	fileServer := file.InitFileServer("./static")

	request, err := requests.ReadRequest(client)

	if err != nil {
		log.Println(err)
		return
	}

	switch request.Path() {
	case "/": 
		if request.Method() == requests.HTTP_METHOD_GET {
			fileBytes := fileServer.Serve("index.html")
			client.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\n\r\n%s\r\n", fileBytes)))
			break
		}
		client.Write([]byte(fmt.Sprintf("HTTP/1.1 405 METHOD NOT ALLOWED\r\n")))
		break
	
	case "/index.html":
		if request.Method() == requests.HTTP_METHOD_GET {
			fileBytes := fileServer.Serve("index.html")
			client.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\n\r\n%s\r\n", fileBytes)))
			break
		}
		client.Write([]byte(fmt.Sprintf("HTTP/1.1 405 METHOD NOT ALLOWED\r\n")))
		break
	
	case "/health":
		if request.Method() == requests.HTTP_METHOD_GET {
			client.Write([]byte("HTTP/1.1 200 OK\r\n"))
			break
		}
		client.Write([]byte(fmt.Sprintf("HTTP/1.1 405 METHOD NOT ALLOWED\r\n")))
		break
	default:
		client.Write([]byte("HTTP/1.1 404 NOT FOUND\r\n"))
	}

	fmt.Printf("Client %s handled successfully\n", client.RemoteAddr())
}
