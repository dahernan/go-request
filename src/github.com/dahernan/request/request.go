package request

import (
	//"encoding/json"
	"bytes"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Request interface {
	Do(method string, endpoint string, requestBody *simplejson.Json) (response *Response, err error)
	Get(endpoint string) (response *Response, err error)
	Post(endpoint string, requestBody *simplejson.Json) (response *Response, err error)
	Put(endpoint string, requestBody *simplejson.Json) (response *Response, err error)
	Delete(endpoint string, requestBody *simplejson.Json) (response *Response, err error)
}

type RequestClient struct {
	httpClient *http.Client
	baseUrl    string
}

type Response struct {
	Json       *simplejson.Json
	StatusCode int
}

func NewRequest(baseUrl string) Request {
	return &RequestClient{baseUrl: baseUrl, httpClient: &http.Client{}}
}

func NewRequestWithClient(baseUrl string, client *http.Client) Request {
	return &RequestClient{baseUrl: baseUrl, httpClient: client}
}

func NewRequestWithTimeout(baseUrl string, timeout time.Duration) Request {
	dialTimeout := func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, timeout)
	}

	transport := http.Transport{
		Dial: dialTimeout,
	}
	client := http.Client{
		Transport: &transport,
	}

	return &RequestClient{baseUrl: baseUrl, httpClient: &client}
}

func (r *RequestClient) Get(endpoint string) (*Response, error) {
	return r.Do("GET", endpoint, nil)
}

func (r *RequestClient) Post(endpoint string, requestBody *simplejson.Json) (*Response, error) {
	return r.Do("POST", endpoint, requestBody)
}

func (r *RequestClient) Put(endpoint string, requestBody *simplejson.Json) (*Response, error) {
	return r.Do("PUT", endpoint, requestBody)
}

func (r *RequestClient) Delete(endpoint string, requestBody *simplejson.Json) (*Response, error) {
	return r.Do("DELETE", endpoint, requestBody)
}

func (r *RequestClient) Do(method string, endpoint string, requestBody *simplejson.Json) (*Response, error) {
	url := r.baseUrl + endpoint

	requestBytes := make([]byte, 0)
	if requestBody != nil {
		var err error
		requestBytes, err = requestBody.Encode()
		if err != nil {
			error := fmt.Errorf("ERROR: [%s|%s]: Can not encode request body: %s\n", method, url, err)
			return nil, error
		}

	}
	clientReq, error := http.NewRequest(method, url, bytes.NewReader(requestBytes))
	if error != nil {
		error := fmt.Errorf("ERROR: [%s|%s]: Can not create http client: %s\n", method, url, error)
		return nil, error
	}

	clientReq.Header.Add("Content-Type", "application/json")
	clientReq.Header.Add("Accept", "application/json")
	response, error := r.httpClient.Do(clientReq)
	status := ""
	if response != nil {
		defer response.Body.Close()
		status = response.Status
	}
	if error != nil {
		error := fmt.Errorf("ERROR: [%s|%s]- [%s]: %s\n", method, url, status, error)
		return nil, error
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		error := fmt.Errorf("ERROR: [%s|%s] - [%s]: Reading the body: %s\n", method, url, status, error)
		return nil, error
	}

	json, error := simplejson.NewJson(body)
	if error != nil {
		error := fmt.Errorf("ERROR: [%s|%s] - [%s]: marshalling json response: %s\nBody--------\n%s\n------------\n", method, url, status, error, body)
		return &Response{Json: nil, StatusCode: response.StatusCode}, error
	}
	// client errors
	if response.StatusCode >= 400 && response.StatusCode <= 499 {
		error := fmt.Errorf("ERROR: [%s|%s] - [%s]: on the resquest \nBody--------\n%s\n------------\n", method, url, status, body)
		return &Response{Json: json, StatusCode: response.StatusCode}, error
	}

	// server errors
	if response.StatusCode >= 500 && response.StatusCode <= 599 {
		error := fmt.Errorf("ERROR: [%s|%s] - [%s]: on the server: %s\nBody--------\n%s\n------------\n", method, url, status, error, body)
		return &Response{Json: nil, StatusCode: response.StatusCode}, error
	}

	return &Response{Json: json, StatusCode: response.StatusCode}, nil
}
