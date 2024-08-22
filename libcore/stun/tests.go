// Copyright 2016 Cong Ding
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stun

import (
	"errors"
	"net"
)

// sendWithLog отправляет запрос с логированием и проверяет ответ
func (c *Client) sendWithLog(conn net.PacketConn, addr *net.UDPAddr, changeIP, changePort bool) (*response, error) {
	c.logger.Debugln("Sending to:", addr)
	resp, err := c.sendBindingReq(conn, addr, changeIP, changePort)
	if err != nil {
		return nil, err
	}
	c.logger.Debugln("Received:", resp)
	if resp == nil && !changeIP && !changePort {
		return nil, errors.New("NAT blocked.")
	}
	if resp != nil && !addrCompare(resp.serverAddr, addr, changeIP, changePort) {
		return nil, errors.New("Server error: response IP/port mismatch")
	}
	return resp, err
}

// addrCompare проверяет, изменились ли IP и порт
func addrCompare(host *Host, addr *net.UDPAddr, IPChange, portChange bool) bool {
	isIPChange := host.IP() != addr.IP.String()
	isPortChange := host.Port() != uint16(addr.Port)
	return isIPChange == IPChange && isPortChange == portChange
}

// test выполняет тест без изменения IP и порта
func (c *Client) test(conn net.PacketConn, addr *net.UDPAddr) (*response, error) {
	return c.sendWithLog(conn, addr, false, false)
}

// testChangePort выполняет тест с изменением порта
func (c *Client) testChangePort(conn net.PacketConn, addr *net.UDPAddr) (*response, error) {
	return c.sendWithLog(conn, addr, false, true)
}

// testChangeBoth выполняет тест с изменением IP и порта
func (c *Client) testChangeBoth(conn net.PacketConn, addr *net.UDPAddr) (*response, error) {
	return c.sendWithLog(conn, addr, true, true)
}

// test1 выполняет тест без изменения IP и порта
func (c *Client) test1(conn net.PacketConn, addr net.Addr) (*response, error) {
	return c.sendBindingReq(conn, addr, false, false)
}

// test2 выполняет тест с изменением IP и порта
func (c *Client) test2(conn net.PacketConn, addr net.Addr) (*response, error) {
	return c.sendBindingReq(conn, addr, true, true)
}

// test3 выполняет тест с изменением порта
func (c *Client) test3(conn net.PacketConn, addr net.Addr) (*response, error) {
	return c.sendBindingReq(conn, addr, false, true)
}
