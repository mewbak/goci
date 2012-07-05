package client

import (
	"bytes"
	"io"
	"net/http"
)

//Codec implements the functionality that a client needs to send requests to
//a service.
type Codec interface {
	ContentType() string
	EncodeRequest(method string, args interface{}) ([]byte, error)
	DecodeResponse(r io.Reader, reply interface{}) error
}

//Client represents an RPC client that can be used to make requests to an rpc
//service.
type Client struct {
	path   string
	codec  Codec
	client *http.Client
}

//New returns a new Client to handle requests to the service at the
//location specified by path. The codec is used to encode and decode the requests
//performed by the given client.
func New(path string, client *http.Client, codec Codec) *Client {
	return &Client{
		path:   path,
		codec:  codec,
		client: client,
	}
}

//Call invokes the named method, waits for it to complete, and returns the results.
func (c *Client) Call(method string, args interface{}, reply interface{}) (err error) {
	//encode our request into a buffer
	buf, err := c.codec.EncodeRequest(method, args)
	if err != nil {
		return
	}
	body := bytes.NewReader(buf)

	//create the post request for the client
	req, err := http.NewRequest("POST", c.path, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", c.codec.ContentType())
	req.Header.Set("Connection", "close") //don't keep it alive

	//invoke the method
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	//read back the response
	err = c.codec.DecodeResponse(resp.Body, reply)
	return
}