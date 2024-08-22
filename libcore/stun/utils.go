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
)

// padToMultipleOfFour дополняет длину среза байтов до ближайшего большего кратного 4.
func padToMultipleOfFour(data []byte) []byte {
	length := uint16(len(data))
	paddingSize := alignToFour(length) - length
	return append(data, make([]byte, paddingSize)...)
}

// alignToFour выравнивает число типа uint16 до ближайшего большего кратного 4.
func alignToFour(n uint16) uint16 {
	return (n + 3) & 0xfffc
}

// isLocalAddress проверяет, является ли localRemote локальным адресом.
func isLocalAddress(local, localRemote string) bool {
	// Сначала разрешаем IP, возвращенный STUN-сервером.
	localRemoteAddr, err := net.ResolveUDPAddr("udp", localRemote)
	if err != nil {
		return false
	}

	// Пытаемся сравнить с локальным адресом на сокете, если он указан.
	localAddr, err := net.ResolveUDPAddr("udp", local)
	if err == nil && localAddr.IP != nil && !localAddr.IP.IsUnspecified() {
		return localAddr.IP.Equal(localRemoteAddr.IP)
	}

	// В случае неудачи проверяем IP всех интерфейсов.
	interfaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}

	for _, addr := range interfaceAddrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		if ip.Equal(localRemoteAddr.IP) {
			return true
		}
	}
	return false
}
