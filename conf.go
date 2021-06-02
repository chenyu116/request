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
	"time"
)

// HTTPTimeout http timeout
type HTTPTimeout struct {
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	HeaderTimeout  time.Duration
	MaxTimeout     time.Duration
}

// Config configure
type Config struct {
	HTTPTimeout        HTTPTimeout // HTTP的超时时间设置
	UseProxy           bool        // 是否使用代理
	ProxyHost          string      // 代理服务器地址
	IsAuthProxy        bool        // 代理服务器是否使用用户认证
	ProxyUser          string      // 代理服务器认证用户名
	ProxyPassword      string      // 代理服务器认证密码
	ReUseTCP           bool        // 为同一地址多次请求复用TCP连接
	InsecureSkipVerify bool        // 忽略证书验证
}

func NewConfig() *Config {
	config := new(Config)

	config.HTTPTimeout.ConnectTimeout = time.Second * 3 // 3s
	config.HTTPTimeout.ReadTimeout = time.Second * 5    // 5s
	config.HTTPTimeout.WriteTimeout = time.Second * 5   // 5s
	config.HTTPTimeout.HeaderTimeout = time.Second * 5  // 5s
	config.HTTPTimeout.MaxTimeout = time.Second * 300   // 300s

	config.UseProxy = false
	config.ProxyHost = ""
	config.IsAuthProxy = false
	config.ProxyUser = ""
	config.ProxyPassword = ""
	config.ReUseTCP = false
	config.InsecureSkipVerify = true

	return config
}
