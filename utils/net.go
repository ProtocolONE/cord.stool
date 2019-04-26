package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"cord.stool/service/models"
	"github.com/pborman/uuid"
)

func GenerateID() string {

	id := uuid.New()
	id = strings.Replace(id, "-", "", -1)
	id = strings.ToUpper(id)
	return id
}

func Get(url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	return httpRequest("GET", url, token, contentType, obj)
}

func Post(url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	return httpRequest("POST", url, token, contentType, obj)
}

func Put(url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	return httpRequest("PUT", url, token, contentType, obj)
}

func Delete(url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	return httpRequest("DELETE", url, token, contentType, obj)
}

func httpRequest(method string, url string, token string, contentType string, obj interface{}) (resp *http.Response, err error) {

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Add("Authorization", token)
	return client.Do(req)
}

func BuldError(r io.Reader) error {

	errorRes := new(models.Error)
	decoder := json.NewDecoder(r)
	if decoder.Decode(&errorRes) == nil {
		return errors.New(errorRes.Message)
	}

	message, _ := ioutil.ReadAll(r)
	return errors.New(string(message))
}
