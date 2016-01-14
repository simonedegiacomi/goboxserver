// Created by Degiacomi Simone

package myhttputil

import (
	"io"
	"fmt"
	"net/http"
)

// Streams upload directly from file -> mime/multipart -> pipe -> http-request
func manageStream (dst *io.PipeWriter, src io.Reader) error {
	
	// Close the stream when the job is done
	defer dst.Close()
	// Copy
	n, err := io.Copy(dst, src)
	
	fmt.Println(n)
	
	if err != nil {
		return err
	}
	
    return err
}

// Creates a new file upload http request with optional extra params
func UploadStream (path, method string,src io.Reader) (*http.Request, error) {
	
	
	ppp, err := http.Post(path, "application/octet-stream", src)
	
	
	// Create the pipe
	//out, in := io.Pipe()

	// Manage the copy of the stream in another go ruotine
		
	//go manageStream(writer, src)
	
	// Return the new request
	
	//req, err := http.NewRequest(method, path, out)
	
	//go func () {
	//	defer in.Close()
	//	io.Copy(in, src)
	//} ()
	
	//fmt.Println(req)
	//fmt.Println(err)
	//req.Header.Set("Content-Type", "application/octet-stream")
	return nil, err
}