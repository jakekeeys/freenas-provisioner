package rest

import (
	"crypto/tls"
	"io"
	"net/http"
)

type Client struct {
	Username string
	Password string
	Host     string
	Client   *http.Client
}

type Interface interface {
	NewRequest(method string, url string, body io.Reader) (*http.Request, error)
	DoRequest(request *http.Request) (*http.Response, error)
}

func New(username, password, host string, insecureSkipVerify bool) Interface {
	client := http.DefaultClient

	if insecureSkipVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &Client{
		Username: username,
		Password: password,
		Host:     host,
		Client:   client,
	}
}

func (c Client) NewRequest(method string, path string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, c.Host+path, body)
	if err != nil {
		return nil, err
	}

	request.SetBasicAuth(c.Username, c.Password)
	request.Header.Add("Content-Type", "application/json")

	return request, nil
}

func (c Client) DoRequest(request *http.Request) (*http.Response, error) {
	response, err := c.Client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}
