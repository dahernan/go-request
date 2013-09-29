package request

import (
	"encoding/json"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/orfjackal/gospec/src/gospec"
	. "github.com/orfjackal/gospec/src/gospec"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
)

func RequestSpec(c gospec.Context) {
	// setup http server for testing

	c.Specify("GET request returns json object", func() {
		ts := httptest.NewServer(http.HandlerFunc(JsonServer(JsonBuilder)))
		url := ts.URL
		defer ts.Close()
		request := NewRequest(url)

		body, _ := request.Get("/")
		c.Expect(body.Get("message").MustString(), Equals, "hello")

	})
	c.Specify("POST request can send and receive json", func() {
		// server that received the request
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c.Expect(req.Method, Equals, "POST")
			c.Expect(req.URL.String(), Equals, "/hello/world")
			c.Expect(req.Header.Get("Accept"), Equals, "application/json")

			body, error := ioutil.ReadAll(req.Body)
			assertNotError(error, c)
			jsonObject, error := simplejson.NewJson(body)
			assertNotError(error, c)
			c.Expect(jsonObject.Get("one").MustString(), Equals, "1 one")
			c.Expect(jsonObject.Get("two").MustString(), Equals, "2 two")
			sendOK(w)
		}))
		url := ts.URL
		defer ts.Close()
		request := NewRequest(url)

		jsonMap := make(map[string]string)
		jsonMap["one"] = "1 one"
		jsonMap["two"] = "2 two"
		jsonBytes, _ := json.Marshal(jsonMap)
		jsonObject, _ := simplejson.NewJson(jsonBytes)

		// method to test
		body, error := request.Post("/hello/world", jsonObject)

		assertNotError(error, c)
		ok, error := body.Get("ok").Bool()
		assertNotError(error, c)
		c.Expect(ok, Equals, true)

	})

	c.Specify("PUT request can send and receive json", func() {
		// server that received the request
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c.Expect(req.Method, Equals, "PUT")
			sendOK(w)
		}))
		url := ts.URL
		defer ts.Close()
		request := NewRequest(url)

		jsonMap := make(map[string]string)
		jsonMap["one"] = "1 one"
		jsonBytes, _ := json.Marshal(jsonMap)
		jsonObject, _ := simplejson.NewJson(jsonBytes)

		// method to test
		body, error := request.Put("/hello/a/put", jsonObject)

		assertNotError(error, c)
		ok, error := body.Get("ok").Bool()
		assertNotError(error, c)
		c.Expect(ok, Equals, true)

	})

	c.Specify("DELETE request can send and receive json", func() {
		// server that received the request
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c.Expect(req.Method, Equals, "DELETE")
			sendOK(w)
		}))
		url := ts.URL
		defer ts.Close()
		request := NewRequest(url)

		jsonMap := make(map[string]string)
		jsonMap["one"] = "1 one"
		jsonBytes, _ := json.Marshal(jsonMap)
		jsonObject, _ := simplejson.NewJson(jsonBytes)

		// method to test
		body, error := request.Delete("/hello/delete", jsonObject)

		assertNotError(error, c)
		ok, error := body.Get("ok").Bool()
		assertNotError(error, c)
		c.Expect(ok, Equals, true)

	})
}

type httpHandlerFunc func(w http.ResponseWriter, req *http.Request)
type jsonHttpBuilderFunc func(req *http.Request) interface{}

func JsonBuilder(req *http.Request) interface{} {
	jsonMap := make(map[string]interface{})
	name := req.URL.Query().Get(":name")
	jsonMap["message"] = "hello" + name
	return jsonMap
}

func sendOK(w http.ResponseWriter) {
	jsonMap := make(map[string]interface{})
	jsonMap["ok"] = true
	jsonBytes, _ := json.Marshal(jsonMap)
	writeJsonBytes(w, jsonBytes)
}

func writeJsonBytes(w http.ResponseWriter, jsonBytes []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(jsonBytes)))
	w.Write(jsonBytes)
}

func JsonServer(builderFunc jsonHttpBuilderFunc) (hanlderFunc httpHandlerFunc) {
	hanlderFunc = func(w http.ResponseWriter, req *http.Request) {
		jsonObject := builderFunc(req)
		jsonBytes, _ := json.Marshal(jsonObject)
		writeJsonBytes(w, jsonBytes)
	}
	return
}

func assertNotError(err interface{}, c gospec.Context) {
	if err != nil {
		fmt.Printf("\tError: %s\n", err)
	}
	c.Expect(err, IsNil)
}
