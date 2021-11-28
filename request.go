/*
   Copyright [2018] [Chen.Yu]

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package request

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
)

type Request struct {
	config *Config
	// request method,default is GET
	method               string
	rawURL               string
	requestBody          io.Reader
	request              *http.Request
	response             *http.Response
	client               *http.Client
	query                url.Values
	header               http.Header
	responseBodyWriteTo  io.ReadWriter
	responseUnwrapType   UnwrapType
	responseUnwrapTarget interface{}
	retryTimes           uint8
	errors               []error
	statusCode           int
}

func Do(rawURL string, options ...Option) error {
	r := &Request{config: NewConfig(), rawURL: rawURL, query: make(url.Values), method: http.MethodGet, header: make(http.Header)}
	for _, option := range options {
		option(r)
	}

	if r.client == nil {
		transport := &http.Transport{
			DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(network, addr, r.config.HTTPTimeout.ConnectTimeout)
				if err != nil {
					return nil, err
				}
				return newTimeoutConn(conn, r.config.HTTPTimeout), nil
			},
			ResponseHeaderTimeout: r.config.HTTPTimeout.HeaderTimeout,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: r.config.InsecureSkipVerify},
		}

		// Proxy
		if r.config.UseProxy && r.config.ProxyHost != "" {
			proxyURL, err := url.Parse(r.config.ProxyHost)
			if err == nil {
				if r.config.IsAuthProxy {
					proxyURL.User = url.UserPassword(r.config.ProxyUser, r.config.ProxyPassword)
				}
				transport.Proxy = http.ProxyURL(proxyURL)
			}
		}
		r.client = &http.Client{Transport: transport}
	}
	return r.Do()
}

func (r *Request) Response() *http.Response {
	return r.response
}

func (r *Request) Request() *http.Request {
	return r.request
}

func (r *Request) StatusCode() int {
	return r.statusCode
}

func (r *Request) Do() (err error) {
	if len(r.errors) > 0 {
		buf := new(bytes.Buffer)
		for _, e := range r.errors {
			buf.WriteString(e.Error())
			buf.WriteString("\n")
		}
		err = errors.New(buf.String())
		return
	}
	URL, err := url.Parse(r.rawURL)
	if err != nil {
		return
	}
	if r.query != nil {
		URL.RawQuery = r.query.Encode()
	}
	var i uint8
	for i = 0; i <= r.retryTimes; i++ {
		r.request, err = http.NewRequestWithContext(context.Background(), r.method, URL.String(), r.requestBody)
		if err != nil {
			continue
		}
		r.request.Header = r.header

		r.response, err = r.client.Do(r.request)
		if err != nil {
			continue
		}
		r.statusCode = r.response.StatusCode
		break
	}
	if err != nil {
		return
	}
	defer r.response.Body.Close()
	if r.responseUnwrapTarget != nil {
		var b []byte
		b, err = ioutil.ReadAll(r.response.Body)
		if err != nil {
			return
		}
		err = json.Unmarshal(b, r.responseUnwrapTarget)
		if err != nil {
			return
		}
	}
	if r.responseBodyWriteTo != nil {
		var b []byte
		b, err = ioutil.ReadAll(r.response.Body)
		if err != nil {
			return
		}
		_, err = r.responseBodyWriteTo.Write(b)
	}
	return
}
