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
	"bytes"
	"encoding/hex"
	"errors"
	"net"
	"time"
)

const (
	numRetransmit  = 9
	defaultTimeout = 100
	maxTimeout     = 1600
	maxPacketSize  = 1024
)

func (c *Client) sendBindingReq(conn net.PacketConn, addr net.Addr, changeIP, changePort bool) (*response, error) {
	// Создание пакета.
	pkt, err := newPacket()
	if err != nil {
		return nil, err
	}
	pkt.types = typeBindingRequest
	softwareAttr := newSoftwareAttribute(c.softwareName)
	pkt.addAttribute(*softwareAttr)

	if changeIP || changePort {
		changeReqAttr := newChangeReqAttribute(changeIP, changePort)
		pkt.addAttribute(*changeReqAttr)
	}

	// Длина атрибута отпечатка должна быть включена в CRC,
	// поэтому мы добавляем его перед вычислением CRC, затем вычитаем после.
	pkt.length += 8
	fingerprintAttr := newFingerprintAttribute(pkt)
	pkt.length -= 8
	pkt.addAttribute(*fingerprintAttr)

	// Отправка пакета.
	return c.send(pkt, conn, addr)
}

// RFC 3489: Клиенты ДОЛЖНЫ повторно передавать запрос, начиная с интервала
// 100 мс, удваивая каждый раз, пока интервал не достигнет 1,6 с.
// Повторные передачи продолжаются с интервалами 1,6 с, пока не будет получен ответ,
// или пока не будет отправлено всего 9 запросов.
func (c *Client) send(pkt *packet, conn net.PacketConn, addr net.Addr) (*response, error) {
	c.logger.Info("\n" + hex.Dump(pkt.bytes()))
	timeout := defaultTimeout
	packetBytes := make([]byte, maxPacketSize)

	for i := 0; i < numRetransmit; i++ {
		// Отправка пакета на сервер.
		length, err := conn.WriteTo(pkt.bytes(), addr)
		if err != nil {
			return nil, err
		}
		if length != len(pkt.bytes()) {
			return nil, errors.New("Ошибка при отправке данных.")
		}

		err = conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Millisecond))
		if err != nil {
			return nil, err
		}

		if timeout < maxTimeout {
			timeout *= 2
		}

		for {
			// Чтение с порта.
			length, raddr, err := conn.ReadFrom(packetBytes)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					break
				}
				return nil, err
			}

			p, err := newPacketFromBytes(packetBytes[:length])
			if err != nil {
				return nil, err
			}

			// Если transId не совпадает, продолжаем чтение, пока не получим
			// совпадающий пакет или не истечет время ожидания.
			if !bytes.Equal(pkt.transID, p.transID) {
				continue
			}

			c.logger.Info("\n" + hex.Dump(packetBytes[:length]))
			resp := newResponse(p, conn)
			resp.serverAddr = newHostFromStr(raddr.String())
			return resp, err
		}
	}
	return nil, nil
}
