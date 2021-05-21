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

type ClientConfig struct {
	DisableSoftware    bool
	DisableFingerprint bool
}

func NewClientConfig() *ClientConfig {
	return &ClientConfig{}
}

// Client is a STUN client, which can be set STUN server address and is used
// to discover NAT type.
type Client struct {
	config *ClientConfig
	// serverAddr   *net.UDPAddr
	softwareName string
	conn         net.PacketConn
	logger       *Logger
}

// NewClient returns a client without network connection. The network
// connection will be build when calling Discover function.
func NewClient(config *ClientConfig) *Client {
	c := new(Client)
	c.config = config
	c.SetSoftwareName(DefaultSoftwareName)
	c.logger = NewLogger()
	return c
}

// NewClientWithConnection returns a client which uses the given connection.
// Please note the connection should be acquired via net.Listen* method.
func NewClientWithConnection(conn net.PacketConn, config *ClientConfig) *Client {
	c := new(Client)
	c.conn = conn
	c.config = config
	c.SetSoftwareName(DefaultSoftwareName)
	c.logger = NewLogger()
	return c
}

// SetVerbose sets the client to be in the verbose mode, which prints
// information in the discover process.
func (c *Client) SetVerbose(v bool) {
	c.logger.SetDebug(v)
}

// SetVVerbose sets the client to be in the double verbose mode, which prints
// information and packet in the discover process.
func (c *Client) SetVVerbose(v bool) {
	c.logger.SetInfo(v)
}

// // SetServerHost allows user to set the STUN hostname and port.
// func (c *Client) SetServerHost(host string, port int) error {
// 	return c.SetServerAddr(net.JoinHostPort(host, strconv.Itoa(port)))
// }

// // SetServerAddr allows user to set the transport layer STUN server address.
// func (c *Client) SetServerAddr(address string) error {
// 	udpAddr, err := net.ResolveUDPAddr("udp", address)
// 	if err != nil {
// 		return err
// 	}
// 	c.serverAddr = udpAddr
// 	return nil
// }

// SetSoftwareName allows user to set the name of the software, which is used
// for logging purpose (NOT used in the current implementation).
func (c *Client) SetSoftwareName(name string) {
	c.softwareName = name
}

func (c *Client) Discover(serverAddr string) (NATType, *Host, error) {
	serverUDPAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return NATError, nil, err
	}
	return c.DiscoverAddr(serverUDPAddr)
}

// DiscoverAddr contacts the STUN server and gets the response of NAT type, host
// for UDP punching.
func (c *Client) DiscoverAddr(serverAddr *net.UDPAddr) (NATType, *Host, error) {
	// if c.serverAddr == nil {
	// 	if err := c.SetServerAddr(DefaultServerAddr); err != nil {
	// 		return NATError, nil, err
	// 	}
	// }
	// serverUDPAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	// if err != nil {
	// 	return NATError, nil, err
	// }
	// Use the connection passed to the client if it is not nil, otherwise
	// create a connection and close it at the end.
	var err error
	conn := c.conn
	if conn == nil {
		conn, err = net.ListenUDP("udp", nil)
		if err != nil {
			return NATError, nil, err
		}
		defer conn.Close()
	}
	return c.discover(conn, serverAddr)
}

func (c *Client) Keepalive(serverAddr string) (*Host, error) {
	serverUDPAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, err
	}
	return c.KeepaliveAddr(serverUDPAddr)
}

// Keepalive sends and receives a bind request, which ensures the mapping stays open
// Only applicable when client was created with a connection.
func (c *Client) KeepaliveAddr(serverAddr *net.UDPAddr) (*Host, error) {
	if c.conn == nil {
		return nil, errors.New("no connection available")
	}
	// if c.serverAddr == nil {
	// 	if err := c.SetServerAddr(DefaultServerAddr); err != nil {
	// 		return nil, err
	// 	}
	// }
	// serverUDPAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	// if err != nil {
	// 	return nil, err
	// }

	resp, err := c.test1(c.conn, serverAddr)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.packet == nil {
		return nil, errors.New("failed to contact")
	}
	return resp.mappedAddr, nil
}
