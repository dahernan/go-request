package request_test

import (
	. "github.com/dahernan/request"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
)

var _ = Describe("Request", func() {
	Describe("Can make basic request sending and receiving JSON", func() {
		It("can do a GET", func() {
			ts := httptest.NewServer(http.HandlerFunc(jsonServer(jsonBuilder)))
			url := ts.URL
			defer ts.Close()
			request := NewRequest(url)

			response, _ := request.Get("/")
			Expect(response.Json.Get("message").MustString()).To(Equal("hello"))
		})

		It("can do a POST", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Method).To(Equal("POST"))
				Expect(req.URL.String()).To(Equal("/hello/world"))
				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
				Expect(req.Header.Get("Accept")).To(Equal("application/json"))

				body, error := ioutil.ReadAll(req.Body)
				Expect(error).To(BeNil())
				jsonObject, error := simplejson.NewJson(body)
				Expect(error).To(BeNil())
				Expect(jsonObject.Get("one").MustString()).To(Equal("1 one"))
				Expect(jsonObject.Get("two").MustString()).To(Equal("2 two"))
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
			response, error := request.Post("/hello/world", jsonObject)

			Expect(error).To(BeNil())
			ok, error := response.Json.Get("ok").Bool()
			Expect(error).To(BeNil())
			Expect(ok).To(BeTrue())
		})

		It("Can do a PUT", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Method).To(Equal("PUT"))
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
			response, error := request.Put("/hello/a/put", jsonObject)

			Expect(error).To(BeNil())
			ok, error := response.Json.Get("ok").Bool()
			Expect(error).To(BeNil())
			Expect(ok).To(BeTrue())

		})

		It("Can do a DELETE", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				Expect(req.Method).To(Equal("DELETE"))
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
			response, error := request.Delete("/hello/delete", jsonObject)

			Expect(error).To(BeNil())
			ok, error := response.Json.Get("ok").Bool()
			Expect(error).To(BeNil())
			Expect(ok).To(BeTrue())

		})

	})

	Describe("Handle client errors", func() {
		It("should hanlde a 404", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				sendNotFound(w)
			}))
			url := ts.URL
			defer ts.Close()
			request := NewRequest(url)
			response, error := request.Get("/")
			//fmt.Printf("Error: %s", error)
			Expect(response.StatusCode).To(Equal(404))
			ok, error := response.Json.Get("exists").Bool()
			Expect(error).To(BeNil())
			Expect(ok).To(BeFalse())
		})

		It("should hanlde a 400", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				sendBadRequest(w)
			}))
			url := ts.URL
			defer ts.Close()
			request := NewRequest(url)
			response, error := request.Get("/")
			fmt.Println(error)
			Expect(response.StatusCode).To(Equal(400))
			Expect(response.Json.Get("error").MustString()).To(Equal("Bad request error"))

		})

	})

	Describe("Handle server errors", func() {
		It("should hanlde a 500", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error on the server"))

			}))
			url := ts.URL
			defer ts.Close()
			request := NewRequest(url)
			response, error := request.Get("/")
			Expect(response.StatusCode).To(Equal(500))
			Expect(error).NotTo(BeNil())
			fmt.Println(error)

		})

	})

	Describe("Handle Json marshalling errors", func() {
		It("should hanlde non json document as an error", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				writeJsonBytes(w, []byte("bad json document"))
			}))
			url := ts.URL
			defer ts.Close()
			request := NewRequest(url)
			response, error := request.Get("/")
			fmt.Println(error)
			Expect(response.Json).To(BeNil())
			Expect(error).NotTo(BeNil())
		})

	})

	Describe("Handle url errors", func() {
		It("should hanlde malformed urls", func() {
			request := NewRequest("wrong")
			response, error := request.Get("/")
			fmt.Println(error)
			Expect(response).To(BeNil())
			Expect(error).NotTo(BeNil())
		})

	})

})

type httpHandlerFunc func(w http.ResponseWriter, req *http.Request)
type jsonHttpBuilderFunc func(req *http.Request) interface{}

func jsonBuilder(req *http.Request) interface{} {
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

func sendNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	jsonMap := make(map[string]interface{})
	jsonMap["exists"] = false
	jsonBytes, _ := json.Marshal(jsonMap)
	writeJsonBytes(w, jsonBytes)
}

func sendBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	jsonMap := make(map[string]interface{})
	jsonMap["error"] = "Bad request error"
	jsonBytes, _ := json.Marshal(jsonMap)
	writeJsonBytes(w, jsonBytes)
}

func writeJsonBytes(w http.ResponseWriter, jsonBytes []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(jsonBytes)))
	w.Write(jsonBytes)
}

func jsonServer(builderFunc jsonHttpBuilderFunc) (hanlderFunc httpHandlerFunc) {
	hanlderFunc = func(w http.ResponseWriter, req *http.Request) {
		jsonObject := builderFunc(req)
		jsonBytes, _ := json.Marshal(jsonObject)
		writeJsonBytes(w, jsonBytes)
	}
	return
}
