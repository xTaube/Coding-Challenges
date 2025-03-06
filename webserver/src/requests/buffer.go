package requests

import (
	"fmt"
	"net"
)

type ReadError struct {
	clientAddr net.Addr
}

func (e *ReadError) Error() string {
	return fmt.Sprintf("Could not read from client: %s", e.clientAddr)
}

type requestBuffer struct {
	buff      []byte
	capacity  int
	readBytes int
}

func (rb *requestBuffer) read(client net.Conn) error {
	n, err := client.Read(rb.buff)
	if err != nil {
		return &ReadError{client.RemoteAddr()}
	}

	rb.readBytes = n

	return nil
}

func (rb *requestBuffer) get() []byte {
	return rb.buff[:rb.readBytes]
}

func initRequestBuffer(buffSize int) *requestBuffer {
	buffer := requestBuffer{make([]byte, buffSize), buffSize, 0}
	return &buffer
}
