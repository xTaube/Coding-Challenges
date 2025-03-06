package http

import (
	"net"
	"strconv"
	"strings"
)

type HTTP_VERSION int

const (
	HTTP_VERSION_1_1 HTTP_VERSION = iota
)


type Response struct {
	httpVersion HTTP_VERSION
	status int
}

func(r *Response) Status() int {
	return r.status
}


func ReadResponse(connection net.Conn) (*Response, error) {
	buff := make([]byte, 1024)
	_, err := connection.Read(buff)
	if err != nil {
		return nil, err
	}

	response, err := parseResponse(buff)

	if err != nil {
		return nil, err
	}

	return response, err
}


func parseResponse(buff []byte) (*Response, error) {
	var response Response

	responseContent := string(buff)

	// parse http version
	idx := strings.IndexByte(responseContent, ' ')
	response.httpVersion = HTTP_VERSION_1_1

	responseContent = responseContent[idx+1:]

	// parse status
	idx = strings.IndexByte(responseContent, ' ')
	status, err := strconv.Atoi(responseContent[:idx])

	if err != nil {
		return nil, err
	}

	response.status = status

	// TODO parse the rest

	return &response, nil
}