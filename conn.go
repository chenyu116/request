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
	"net"
	"time"
)

// Handle http timeout
type timeoutConn struct {
	conn    net.Conn
	timeout HTTPTimeout
}

func newTimeoutConn(conn net.Conn, timeout HTTPTimeout) *timeoutConn {
	if timeout.MaxTimeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(timeout.MaxTimeout))
	}
	return &timeoutConn{
		conn:    conn,
		timeout: timeout,
	}
}

func (c *timeoutConn) Read(b []byte) (n int, err error) {
	if c.timeout.ReadTimeout > 0 {
		_ = c.SetReadDeadline(time.Now().Add(c.timeout.ReadTimeout))
	}
	n, err = c.conn.Read(b)
	if c.timeout.MaxTimeout > 0 {
		_ = c.SetReadDeadline(time.Now().Add(c.timeout.MaxTimeout))
	}
	return n, err
}

func (c *timeoutConn) Write(b []byte) (n int, err error) {
	if c.timeout.WriteTimeout > 0 {
		_ = c.SetWriteDeadline(time.Now().Add(c.timeout.WriteTimeout))
	}
	n, err = c.conn.Write(b)
	if c.timeout.MaxTimeout > 0 {
		_ = c.SetWriteDeadline(time.Now().Add(c.timeout.MaxTimeout))
	}
	return n, err
}

func (c *timeoutConn) Close() error {
	return c.conn.Close()
}

func (c *timeoutConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *timeoutConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *timeoutConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *timeoutConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *timeoutConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
