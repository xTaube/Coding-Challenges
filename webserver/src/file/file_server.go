package file

import (
	"fmt"
	"io"
	"os"
)


type FileServer struct {
	dir string
}

func(fileServer *FileServer) Serve(file string) []byte {
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", fileServer.dir, file), os.O_RDONLY, 0)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	fileBytes, _ := io.ReadAll(f)

	return fileBytes
}

func InitFileServer(dir string) *FileServer {
	return &FileServer{dir}
}