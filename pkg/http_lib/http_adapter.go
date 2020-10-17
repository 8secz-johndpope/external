package http_lib

import (
	"bytes"
	"fmt"
	"net/http"
)

func Post(url string, body []byte, headers map[string]string) (*http.Response, error) {
	return HttpRequest(url, body, headers, "POST")
}

func Get(url string, body []byte, headers map[string]string) (*http.Response, error) {
	return HttpRequest(url, body, headers, "GET")
}

func Put(url string, body []byte, headers map[string]string) (*http.Response, error) {
	return HttpRequest(url, body, headers, "PUT")
}

func Patch(url string, body []byte, headers map[string]string) (*http.Response, error) {
	return HttpRequest(url, body, headers, "PATCH")
}

func Delete(url string, body []byte, headers map[string]string) (*http.Response, error) {
	return HttpRequest(url, body, headers, "DELETE")
}

//to avoid duplication in other functions
func HttpRequest(url string, body []byte, headers map[string]string, method string) (*http.Response, error) {
	req, reqErr := http.NewRequest(method, url, bytes.NewBuffer(body))

	if reqErr != nil {
		return nil, reqErr
	}

	if headers != nil && len(headers) > 0 {
		for k, v := range headers {
			fmt.Println("k:", k, "v:", v)
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
	resp, resErr := client.Do(req)

	if resErr != nil {
		return nil, resErr
	}

	return resp, nil
}
