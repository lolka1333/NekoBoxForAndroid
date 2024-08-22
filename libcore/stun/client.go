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
	"strconv"
)

// Client represents a STUN client, which can set the STUN server address and is used
// to discover the NAT type.
type Client struct {
	serverAddr   string
	softwareName string
	conn         net.PacketConn
	logger       *Logger
}

// NewClient returns a client without a network connection. The network
// connection will be established when calling the Discover function.
func NewClient() *Client {
	client := &Client{}
	client.SetSoftwareName(DefaultSoftwareName)
	client.logger = NewLogger()
	return client
}

// NewClientWithConnection returns a client that uses the given connection.
// Note that the connection should be acquired via the net.Listen* method.
func NewClientWithConnection(conn net.PacketConn) *Client {
	client := &Client{conn: conn}
	client.SetSoftwareName(DefaultSoftwareName)
	client.logger = NewLogger()
	return client
}

// SetVerbose sets the client to verbose mode, which prints
// information during the discovery process.
func (c *Client) SetVerbose(verbose bool) {
	c.logger.SetDebug(verbose)
}

// SetVVerbose sets the client to double verbose mode, which prints
// information and packets during the discovery process.
func (c *Client) SetVVerbose(verbose bool) {
	c.logger.SetInfo(verbose)
}

// SetServerHost allows the user to set the STUN hostname and port.
func (c *Client) SetServerHost(host string, port int) {
	c.serverAddr = net.JoinHostPort(host, strconv.Itoa(port))
}

// SetServerAddr allows the user to set the transport layer STUN server address.
func (c *Client) SetServerAddr(address string) {
	c.serverAddr = address
}

// SetSoftwareName allows the user to set the name of the software, which is used
// for logging purposes (NOT used in the current implementation).
func (c *Client) SetSoftwareName(name string) {
	c.softwareName = name
}

// Discover contacts the STUN server and gets the response of NAT type, host
// for UDP punching.
func (c *Client) Discover() (NATType, *Host, error, bool) {
	if c.serverAddr == "" {
		c.SetServerAddr(DefaultServerAddr)
	}
	serverUDPAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	if err != nil {
		return NATError, nil, err, false
	}
	// Use the connection passed to the client if it is not nil, otherwise
	// create a connection and close it at the end.
	conn := c.conn
	if conn == nil {
		conn, err = net.ListenUDP("udp", nil)
		if err != nil {
			return NATError, nil, err, false
		}
		defer conn.Close()
	}
	return c.discover(conn, serverUDPAddr)
}

// BehaviorTest performs a NAT behavior test.
func (c *Client) BehaviorTest() (*NATBehavior, error) {
	if c.serverAddr == "" {
		c.SetServerAddr(DefaultServerAddr)
	}
	serverUDPAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	if err != nil {
		return nil, err
	}
	// Use the connection passed to the client if it is not nil, otherwise
	// create a connection and close it at the end.
	conn := c.conn
	if conn == nil {
		conn, err = net.ListenUDP("udp", nil)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
	}
	return c.behaviorTest(conn, serverUDPAddr)
}

// Keepalive sends and receives a bind request, which ensures the mapping stays open.
// Only applicable when the client was created with a connection.
func (c *Client) Keepalive() (*Host, error) {
	if c.conn == nil {
		return nil, errors.New("no connection available")
	}
	if c.serverAddr == "" {
		c.SetServerAddr(DefaultServerAddr)
	}
	serverUDPAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	if err != nil {
		return nil, err
	}

	resp, err := c.test1(c.conn, serverUDPAddr)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.packet == nil {
		return nil, errors.New("failed to contact")
	}
	return resp.mappedAddr, nil
}
