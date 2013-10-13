package request

import (
	//"encoding/json"
	//"fmt"
	"bytes"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
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
		requestBytes, _ = requestBody.Encode()
	}
	clientReq, _ := http.NewRequest(method, url, bytes.NewReader(requestBytes))

	clientReq.Header.Add("Content-Type", "application/json")
	clientReq.Header.Add("Accept", "application/json")
	response, error := r.httpClient.Do(clientReq)
	defer response.Body.Close()

	if error != nil {
		log.Printf("Error %s %s with message %s", method, url, error)
		return nil, error
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		log.Printf("Error reading response Body %s\n for the url %s with message: %s", response.Body, url, error)
		return nil, error
	}

	json, _ := simplejson.NewJson(body)
	// client errors
	if response.StatusCode >= 400 && response.StatusCode <= 499 {
		error := fmt.Errorf("Error on response for the url %s, with message,\n\t%s\n\t%s", url, response.Status, body)
		return &Response{Json: json, StatusCode: response.StatusCode}, error
	}

	// server errors
	if response.StatusCode >= 500 && response.StatusCode <= 599 {
		return &Response{Json: nil, StatusCode: response.StatusCode}, fmt.Errorf("Error on response for the url %s, with message,\n\t%s\n\t%s", url, response.Status, body)
	}

	return &Response{Json: json, StatusCode: response.StatusCode}, nil
}
