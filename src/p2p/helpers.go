// Copyright 2018 The Gringo Developers. All rights reserved.
// Use of this source code is governed by a GNU GENERAL PUBLIC LICENSE v3
// license that can be found in the LICENSE file.

package p2p

import (
	"io"
	"net"
	"github.com/sirupsen/logrus"
	"encoding/binary"
	"errors"
	"fmt"
)

// serializeTCPAddr helps to serialize net.TCPAddr
func serializeTCPAddr(buff io.Writer, addr *net.TCPAddr) {
	IP := addr.IP.To4()
	if IP == nil {
		IP = addr.IP
	}

	switch len(IP) {
	case net.IPv4len:
		{
			if _, err := buff.Write([]byte{0}); err != nil {
				logrus.Fatal(err)
			}

			if _, err := buff.Write(IP); err != nil {
				logrus.Fatal(err)
			}
		}
	case net.IPv6len:
		{
			if _, err := buff.Write([]byte{1}); err != nil {
				logrus.Fatal(err)
			}

			for i := 0; i < 8; i += 2 {
				segment := (uint16(IP[i]) << 8) + uint16(IP[i + 1])

				if err := binary.Write(buff, binary.BigEndian, segment); err != nil {
					logrus.Fatal(err)
				}
			}
		}
	default:
		logrus.Fatal("invalid netaddr")
	}

	if err := binary.Write(buff, binary.BigEndian, uint16(addr.Port)); err != nil {
		logrus.Fatal(err)
	}
}

// deserializeTCPAddr helps to deserialize net.TCPAddr
func deserializeTCPAddr(r io.Reader) (*net.TCPAddr, error) {
	var ipFlag int8
	var ipAddr []byte
	var ipPort uint16

	if err := binary.Read(r, binary.BigEndian, &ipFlag); err != nil {
		return nil, err
	}

	switch ipFlag {
	case 0: // for ipv4 addr
		ipAddr = make([]byte, net.IPv4len)

		if _, err := io.ReadFull(r, ipAddr); err != nil {
			return nil, err
		}
	case 1: // for ipv6 addr
		ipAddr = make([]byte, net.IPv6len)

		for i := 0; i < 8; i += 2 {
			var segment uint16

			if err := binary.Read(r, binary.BigEndian, segment); err != nil {
				return nil, err
			}

			ipAddr[i] = byte(segment >> 8)
			ipAddr[i + 1] = byte(segment)
		}

	default:
		return nil, fmt.Errorf("invalid ipFlag: %V", ipFlag)
	}

	if err := binary.Read(r, binary.BigEndian, &ipPort); err != nil {
		return nil, err
	}

	return  &net.TCPAddr{
		IP: ipAddr,
		Port: int(ipPort),
	}, nil
}