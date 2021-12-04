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
	"fmt"
	"io"
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
	non20xIsError        bool
}

func POST(rawURL string, options ...Option) (*Request, error) {
	options = append(options, WithMethodPost())
	return Do(rawURL, options...)
}

func PUT(rawURL string, options ...Option) (*Request, error) {
	options = append(options, WithMethodPut())
	return Do(rawURL, options...)
}

func PATCH(rawURL string, options ...Option) (*Request, error) {
	options = append(options, WithMethodPatch())
	return Do(rawURL, options...)
}
func OPTIONS(rawURL string, options ...Option) (*Request, error) {
	options = append(options, WithMethodOptions())
	return Do(rawURL, options...)
}

func DELETE(rawURL string, options ...Option) (*Request, error) {
	options = append(options, WithMethodDelete())
	return Do(rawURL, options...)
}

func TRACE(rawURL string, options ...Option) (*Request, error) {
	options = append(options, WithMethodTrace())
	return Do(rawURL, options...)
}

func HEAD(rawURL string, options ...Option) (*Request, error) {
	options = append(options, WithMethodHead())
	return Do(rawURL, options...)
}
func CONNECT(rawURL string, options ...Option) (*Request, error) {
	options = append(options, WithMethodConnect())
	return Do(rawURL, options...)
}

// NewRequest deprecated method
func NewRequest(rawURL string, options ...Option) *Request {
	r := &Request{config: NewConfig(), rawURL: rawURL, query: make(url.Values), method: http.MethodGet, header: make(http.Header), non20xIsError: true}
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
	return r
}

func Do(rawURL string, options ...Option) (*Request, error) {
	r := &Request{config: NewConfig(), rawURL: rawURL, query: make(url.Values), method: http.MethodGet, header: make(http.Header), non20xIsError: true}
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
	return r, r.send()
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

func (r *Request) Do() (statusCode int, err error) {
	URL, err := url.Parse(r.rawURL)
	if err != nil {
		return
	}
	if r.query != nil {
		URL.RawQuery = r.query.Encode()
	}
	var i uint8
	for i = 0; i <= r.retryTimes; i++ {
		r.request, err = http.NewRequest(r.method, URL.String(), r.requestBody)
		if err != nil {
			continue
		}
		r.request.Header = r.header

		r.response, err = r.client.Do(r.request)
		if err != nil {
			continue
		}
		statusCode = r.response.StatusCode
		break
	}
	if err != nil {
		return
	}
	defer r.response.Body.Close()
	if r.non20xIsError && statusCode >= 300 {
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, r.response.Body)
		if err != nil {
			return
		}
		err = fmt.Errorf("%s", buf.String())
		return
	}
	if statusCode != http.StatusNoContent {
		if r.responseUnwrapTarget != nil || r.responseBodyWriteTo != nil {
			if r.responseUnwrapTarget == nil && r.responseBodyWriteTo != nil {
				_, err = io.Copy(r.responseBodyWriteTo, r.response.Body)
				if err != nil {
					return
				}
			} else if r.responseUnwrapTarget != nil && r.responseBodyWriteTo == nil {
				err = json.NewDecoder(r.response.Body).Decode(r.responseUnwrapTarget)
				if err != nil {
					return
				}
			} else {
				buf := new(bytes.Buffer)
				_, err = io.Copy(buf, r.response.Body)
				if err != nil {
					return
				}
				err = json.Unmarshal(buf.Bytes(), r.responseUnwrapTarget)
				if err != nil {
					return
				}
				_, err = io.Copy(r.responseBodyWriteTo, buf)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

func (r *Request) send() (err error) {
	URL, err := url.Parse(r.rawURL)
	if err != nil {
		return
	}
	if r.query != nil {
		URL.RawQuery = r.query.Encode()
	}
	var i uint8
	for i = 0; i <= r.retryTimes; i++ {
		r.request, err = http.NewRequest(r.method, URL.String(), r.requestBody)
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
	if r.non20xIsError && r.statusCode >= 300 {
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, r.response.Body)
		if err != nil {
			return
		}
		err = fmt.Errorf("%s", buf.String())
		return
	}
	if r.statusCode != http.StatusNoContent {
		if r.responseUnwrapTarget != nil || r.responseBodyWriteTo != nil {
			if r.responseUnwrapTarget == nil && r.responseBodyWriteTo != nil {
				_, err = io.Copy(r.responseBodyWriteTo, r.response.Body)
				if err != nil {
					return
				}
			} else if r.responseUnwrapTarget != nil && r.responseBodyWriteTo == nil {
				err = json.NewDecoder(r.response.Body).Decode(r.responseUnwrapTarget)
				if err != nil {
					return
				}
			} else {
				buf := new(bytes.Buffer)
				_, err = io.Copy(buf, r.response.Body)
				if err != nil {
					return
				}
				err = json.Unmarshal(buf.Bytes(), r.responseUnwrapTarget)
				if err != nil {
					return
				}
				_, err = io.Copy(r.responseBodyWriteTo, buf)
				if err != nil {
					return
				}
			}
		}
	}

	return
}
