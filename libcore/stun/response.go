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
	"fmt"
	"net"
)

type response struct {
	packet      *packet // Original packet from the server
	serverAddr  *Host   // Address from which the packet was received
	changedAddr *Host   // Parsed from packet
	mappedAddr  *Host   // Parsed from packet, external address of client NAT
	otherAddr   *Host   // Parsed from packet, to replace changedAddr in RFC 5780
	identical   bool    // True if mappedAddr is in local address list
}

func newResponse(pkt *packet, conn net.PacketConn) *response {
	resp := &response{
		packet:      pkt,
		serverAddr:  nil,
		changedAddr: nil,
		mappedAddr:  nil,
		otherAddr:   nil,
		identical:   false,
	}

	if pkt == nil {
		return resp
	}

	// RFC 3489 doesn't require the server to return XOR mapped address.
	mappedAddr := pkt.getXorMappedAddr()
	if mappedAddr == nil {
		mappedAddr = pkt.getMappedAddr()
	}
	resp.mappedAddr = mappedAddr

	// Determine if the mapped address is identical to a local address
	localAddrStr := conn.LocalAddr().String()
	if mappedAddr != nil {
		mappedAddrStr := mappedAddr.String()
		resp.identical = isLocalAddress(localAddrStr, mappedAddrStr)
	}

	// Compute changedAddr
	changedAddr := pkt.getChangedAddr()
	if changedAddr != nil {
		resp.changedAddr = newHostFromStr(changedAddr.String())
	}

	// Compute otherAddr
	otherAddr := pkt.getOtherAddr()
	if otherAddr != nil {
		resp.otherAddr = newHostFromStr(otherAddr.String())
	}

	return resp
}

// String is used only for verbose mode output.
func (r *response) String() string {
	if r == nil {
		return "Nil"
	}
	return fmt.Sprintf("{packet nil: %v, local: %v, remote: %v, changed: %v, other: %v, identical: %v}",
		r.packet == nil,
		r.mappedAddr,
		r.serverAddr,
		r.changedAddr,
		r.otherAddr,
		r.identical)
}
