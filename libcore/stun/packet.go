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
	"crypto/rand"
	"encoding/binary"
	"errors"
	"math"
)

type packet struct {
	types      uint16
	length     uint16
	transID    []byte // 4 байта magic cookie + 12 байт идентификатор транзакции
	attributes []attribute
}

func newPacket() (*packet, error) {
	pkt := &packet{
		transID:    make([]byte, 16),
		attributes: make([]attribute, 0, 10),
		length:     0,
	}
	binary.BigEndian.PutUint32(pkt.transID[:4], magicCookie)
	if _, err := rand.Read(pkt.transID[4:]); err != nil {
		return nil, err
	}
	return pkt, nil
}

func newPacketFromBytes(packetBytes []byte) (*packet, error) {
	const headerSize = 20
	if len(packetBytes) < headerSize {
		return nil, errors.New("received data length too short")
	}
	if len(packetBytes) > math.MaxUint16+headerSize {
		return nil, errors.New("received data length too long")
	}

	pkt := &packet{
		types:      binary.BigEndian.Uint16(packetBytes[0:2]),
		length:     binary.BigEndian.Uint16(packetBytes[2:4]),
		transID:    packetBytes[4:20],
		attributes: make([]attribute, 0, 10),
	}

	packetBytes = packetBytes[headerSize:]
	for pos := uint16(0); pos+4 < uint16(len(packetBytes)); {
		types := binary.BigEndian.Uint16(packetBytes[pos : pos+2])
		length := binary.BigEndian.Uint16(packetBytes[pos+2 : pos+4])
		end := pos + 4 + length
		if end < pos+4 || end > uint16(len(packetBytes)) {
			return nil, errors.New("received data format mismatch")
		}
		value := packetBytes[pos+4 : end]
		attribute := newAttribute(types, value)
		pkt.addAttribute(*attribute)
		pos += align(length) + 4
	}
	return pkt, nil
}

func (p *packet) addAttribute(attr attribute) {
	p.attributes = append(p.attributes, attr)
	p.length += align(attr.length) + 4
}

func (p *packet) bytes() []byte {
	packetBytes := make([]byte, 4)
	binary.BigEndian.PutUint16(packetBytes[0:2], p.types)
	binary.BigEndian.PutUint16(packetBytes[2:4], p.length)
	packetBytes = append(packetBytes, p.transID...)
	for _, attr := range p.attributes {
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, attr.types)
		packetBytes = append(packetBytes, buf...)
		binary.BigEndian.PutUint16(buf, attr.length)
		packetBytes = append(packetBytes, buf...)
		packetBytes = append(packetBytes, attr.value...)
	}
	return packetBytes
}

func (p *packet) getSourceAddr() *Host {
	return p.getRawAddr(attributeSourceAddress)
}

func (p *packet) getMappedAddr() *Host {
	return p.getRawAddr(attributeMappedAddress)
}

func (p *packet) getChangedAddr() *Host {
	return p.getRawAddr(attributeChangedAddress)
}

func (p *packet) getOtherAddr() *Host {
	return p.getRawAddr(attributeOtherAddress)
}

func (p *packet) getRawAddr(attrType uint16) *Host {
	for _, attr := range p.attributes {
		if attr.types == attrType {
			return attr.rawAddr()
		}
	}
	return nil
}

func (p *packet) getXorMappedAddr() *Host {
	addr := p.getXorAddr(attributeXorMappedAddress)
	if addr == nil {
		addr = p.getXorAddr(attributeXorMappedAddressExp)
	}
	return addr
}

func (p *packet) getXorAddr(attrType uint16) *Host {
	for _, attr := range p.attributes {
		if attr.types == attrType {
			return attr.xorAddr(p.transID)
		}
	}
	return nil
}
