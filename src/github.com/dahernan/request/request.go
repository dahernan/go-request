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
	Do(method string, endpoint string, requestBody *simplejson.Json) (json *simplejson.Json, err error)
	Get(endpoint string) (json *simplejson.Json, err error)
	Post(endpoint string, requestBody *simplejson.Json) (json *simplejson.Json, err error)
	Put(endpoint string, requestBody *simplejson.Json) (json *simplejson.Json, err error)
	Delete(endpoint string, requestBody *simplejson.Json) (json *simplejson.Json, err error)
}

type RequestClient struct {
	httpClient *http.Client
	baseUrl    string
}

func NewRequest(baseUrl string) Request {
	return &RequestClient{baseUrl: baseUrl, httpClient: &http.Client{}}
}

func (r *RequestClient) Get(endpoint string) (*simplejson.Json, error) {
	return r.Do("GET", endpoint, nil)
}

func (r *RequestClient) Post(endpoint string, requestBody *simplejson.Json) (*simplejson.Json, error) {
	return r.Do("POST", endpoint, requestBody)
}

func (r *RequestClient) Put(endpoint string, requestBody *simplejson.Json) (*simplejson.Json, error) {
	return r.Do("PUT", endpoint, requestBody)
}

func (r *RequestClient) Delete(endpoint string, requestBody *simplejson.Json) (*simplejson.Json, error) {
	return r.Do("DELETE", endpoint, requestBody)
}

func (r *RequestClient) Do(method string, endpoint string, requestBody *simplejson.Json) (*simplejson.Json, error) {
	url := r.baseUrl + endpoint

	requestBytes := make([]byte, 0)
	if requestBody != nil {
		requestBytes, _ = requestBody.Encode()
	}
	clientReq, _ := http.NewRequest(method, url, bytes.NewReader(requestBytes))

	clientReq.Header.Add("Content-Type", "application/json")
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

	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		return nil, fmt.Errorf("Error on response for the url %s, with message,\n\t%s\n\t%s", url, response.Status, body)
	}

	json, _ := simplejson.NewJson(body)

	return json, nil
}
