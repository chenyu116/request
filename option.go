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
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type UnwrapType string

const (
	UnwrapTypeJson       UnwrapType = "Json"
	UnwrapTypeReadWriter UnwrapType = "ReadWriter"
)

type Option func(request *Request)

func WithConfig(config *Config) Option {
	return func(request *Request) {
		request.config = config
	}
}

func WithRetryTimes(retryTimes uint8) Option {
	return func(request *Request) {
		request.retryTimes = retryTimes
	}
}

func WithMethodPost() Option {
	return func(request *Request) {
		request.method = http.MethodPost
	}
}

func WithMethodPut() Option {
	return func(request *Request) {
		request.method = http.MethodPut
	}
}

func WithMethodHead() Option {
	return func(request *Request) {
		request.method = http.MethodHead
	}
}

func WithMethodConnect() Option {
	return func(request *Request) {
		request.method = http.MethodConnect
	}
}

func WithMethodPatch() Option {
	return func(request *Request) {
		request.method = http.MethodPatch
	}
}

func WithMethodDelete() Option {
	return func(request *Request) {
		request.method = http.MethodDelete
	}
}

func WithMethodTrace() Option {
	return func(request *Request) {
		request.method = http.MethodTrace
	}
}

func WithMethodOptions() Option {
	return func(request *Request) {
		request.method = http.MethodOptions
	}
}

func WithClient(client *http.Client) Option {
	return func(request *Request) {
		request.client = client
	}
}

func WithResponseBodyWriteTo(responseBodyWriteTo io.ReadWriter) Option {
	return func(request *Request) {
		request.responseUnwrapType = UnwrapTypeReadWriter
		request.responseBodyWriteTo = responseBodyWriteTo
	}
}
func WithResponseBodyToJson(unwrapTarget interface{}) Option {
	return func(request *Request) {
		request.responseUnwrapType = UnwrapTypeJson
		request.responseUnwrapTarget = unwrapTarget
	}
}

func WithBodyForm(keyPairs ...string) Option {
	return func(request *Request) {
		v := url.Values{}
		keyPairsLen := len(keyPairs)
		if keyPairsLen%2 == 1 {
			keyPairs = append(keyPairs, "")
			keyPairsLen++
		}
		for i := 0; i < keyPairsLen; i += 2 {
			if keyPairs[i] == "" {
				continue
			}
			v.Set(keyPairs[i], keyPairs[i+1])
		}
		request.header.Set("content-type", "application/x-www-form-urlencoded")
		request.requestBody = strings.NewReader(v.Encode())
	}
}

func WithBodyJson(body interface{}, escape ...bool) Option {
	return func(request *Request) {
		buf := new(bytes.Buffer)
		encoder := json.NewEncoder(buf)
		escapeHTML := false
		if len(escape) > 0 && escape[0] == true {
			escapeHTML = true
		}
		encoder.SetEscapeHTML(escapeHTML)
		err := encoder.Encode(body)
		if err != nil {
			request.errors = append(request.errors, err)
			return
		}
		request.header.Set("content-type", "application/json; charset=utf-8")
		request.requestBody = buf
	}
}

type File struct {
	FieldName string
	Path      string
}

func WithBodyFiles(files []File, fieldsKeyPairs ...string) Option {
	return func(request *Request) {
		if len(files) == 0 {
			request.errors = append(request.errors, errors.New("no file found"))
			return
		}
		body := new(bytes.Buffer)
		// 文件写入 body
		writer := multipart.NewWriter(body)
		for _, f := range files {
			file, err := os.Open(f.Path)
			if err != nil {
				request.errors = append(request.errors, err)
				return
			}
			part, err := writer.CreateFormFile(f.FieldName, filepath.Base(f.Path))
			if err != nil {
				_ = file.Close()
				request.errors = append(request.errors, err)
				return
			}
			_, err = io.Copy(part, file)
			if err != nil {
				_ = file.Close()
				request.errors = append(request.errors, err)
				return
			}
			_ = file.Close()
		}
		fieldsLen := len(fieldsKeyPairs)
		if fieldsLen%2 == 1 {
			fieldsKeyPairs = append(fieldsKeyPairs, "")
		}
		for i := 0; i < fieldsLen; i += 2 {
			if fieldsKeyPairs[i] == "" {
				continue
			}
			err := writer.WriteField(fieldsKeyPairs[i], fieldsKeyPairs[i+1])
			if err != nil {
				request.errors = append(request.errors, err)
				return
			}
		}
		if err := writer.Close(); err != nil {
			request.errors = append(request.errors, err)
			return
		}
		request.header.Set("content-type", writer.FormDataContentType())
		request.requestBody = body
	}
}

func WithQuery(keyPairs ...string) Option {
	return func(request *Request) {
		keyPairsLen := len(keyPairs)
		if keyPairsLen%2 == 1 {
			keyPairs = append(keyPairs, "")
			keyPairsLen++
		}

		for i := 0; i < keyPairsLen; i += 2 {
			if keyPairs[i] == "" {
				continue
			}
			request.query.Set(keyPairs[i], keyPairs[i+1])
		}
	}
}

func WithBasicAuth(username, password string) Option {
	return func(request *Request) {
		buf := new(bytes.Buffer)
		buf.WriteString(username)
		buf.WriteString(":")
		buf.WriteString(password)
		encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())
		buf.Reset()
		buf.WriteString("Basic ")
		buf.WriteString(encodedString)
		request.header.Set("Authorization", buf.String())
	}
}
func WithHeader(keyPairs ...string) Option {
	return func(request *Request) {
		keyPairsLen := len(keyPairs)
		if keyPairsLen%2 == 1 {
			keyPairs = append(keyPairs, "")
			keyPairsLen++
		}
		for i := 0; i < keyPairsLen; i += 2 {
			if keyPairs[i] == "" {
				continue
			}
			request.header.Set(keyPairs[i], keyPairs[i+1])
		}
	}
}
