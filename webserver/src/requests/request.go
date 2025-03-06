package requests

import (
	"fmt"
	"net"
	"strings"
)

type HTTP_METHOD int

const (
	HTTP_METHOD_GET HTTP_METHOD = iota
	HTTP_METHOD_POST

	HTTP_METHOD_UNKNOWN
)

type HttpMethodUnknownError struct {
	method string
}

func(err *HttpMethodUnknownError) Error() string {
	return fmt.Sprintf("Method %s is unknown", err.method)
}

type HTTP_VERSION int

const (
	HTTP_VERSION_1_1 HTTP_VERSION = iota

	HTTP_VERSION_UNKNOWN
)

type Request struct {
	httpMethod HTTP_METHOD
	httpVersion HTTP_VERSION
	path string
}

func(r *Request) Path() string {
	return r.path
}

func(r *Request) Method() HTTP_METHOD {
	return r.httpMethod
}

func ReadRequest(client net.Conn) (*Request, error) {
	buffer := initRequestBuffer(1024)
	err := buffer.read(client)
	if err != nil {
		return nil, err
	}

	request , err := parseRequest(buffer)
	return request, err
}

func parseHttpMethod(httpMethod string) HTTP_METHOD {
	switch httpMethod {
	case "GET":
		return HTTP_METHOD_GET
	case "POST":
		return HTTP_METHOD_POST
	default:
		return HTTP_METHOD_UNKNOWN
	}
}

func parseRequest(buffer *requestBuffer) (*Request, error) {
	var request Request

	requestContent := string(buffer.get())

	// parse http method
	idx := strings.IndexByte(requestContent, ' ')
	
	httpMethodString := requestContent[:idx]
	httpMethod := parseHttpMethod(httpMethodString)
	
	if httpMethod == HTTP_METHOD_UNKNOWN {
		return nil, &HttpMethodUnknownError{httpMethodString}
	}
 
	requestContent = requestContent[idx+1:]
	
	// parse request path
	idx = strings.IndexByte(requestContent, ' ')
	request.path = requestContent[:idx]

	requestContent = requestContent[idx+1:]

	// parse http version
	idx = strings.IndexByte(requestContent, '\r')
	request.httpVersion = HTTP_VERSION_1_1

	requestContent = requestContent[idx+1:]

	return &request, nil
}
