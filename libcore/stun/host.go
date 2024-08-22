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
	"net"
	"strconv"
)

// Host представляет сетевой адрес, включая семейство адресов, IP-адрес и порт.
type Host struct {
	family uint16
	ip     string
	port   uint16
}

// newHostFromStr создает новый объект Host из строки.
func newHostFromStr(address string) *Host {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil
	}

	host := &Host{
		ip:   udpAddr.IP.String(),
		port: uint16(udpAddr.Port),
	}

	if udpAddr.IP.To4() != nil {
		host.family = attributeFamilyIPv4
	} else {
		host.family = attributeFamilyIPV6
	}

	return host
}

// Family возвращает тип семейства адресов хоста (IPv4 или IPv6).
func (h *Host) Family() uint16 {
	return h.family
}

// IP возвращает IP-адрес хоста.
func (h *Host) IP() string {
	return h.ip
}

// Port возвращает номер порта хоста.
func (h *Host) Port() uint16 {
	return h.port
}

// TransportAddr возвращает адрес транспортного уровня хоста.
func (h *Host) TransportAddr() string {
	return net.JoinHostPort(h.ip, strconv.Itoa(int(h.port)))
}

// String возвращает строковое представление адреса хоста.
func (h *Host) String() string {
	return h.TransportAddr()
}
