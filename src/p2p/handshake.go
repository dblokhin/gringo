package p2p

import (
	"net"
	"consensus"
	"encoding/binary"
	"bytes"
	"github.com/sirupsen/logrus"
	"io"
	"errors"
)

// First part of a handshake, sender advertises its version and
// characteristics.
type hand struct {
	// protocol version of the sender
	Version uint32
	// Capabilities of the sender
	Capabilities consensus.Capabilities
	// randomly generated for each handshake, helps detect self
	Nonce uint64
	// total difficulty accumulated by the sender, used to check whether sync
	// may be needed
	TotalDifficulty consensus.Difficulty
	// network address of the sender
	SenderAddr   *net.TCPAddr
	ReceiverAddr *net.TCPAddr

	// name of version of the software
	UserAgent string
}

func (h *hand) Bytes() []byte {
	logrus.Info("hand struct to bytes")
	buff := new(bytes.Buffer)

	if err := binary.Write(buff, binary.BigEndian, h.Version); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint32(h.Capabilities)); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, h.Nonce); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(h.TotalDifficulty)); err != nil {
		logrus.Fatal(err)
	}

	// Write Sender addr
	switch len(h.SenderAddr.IP) {
	case net.IPv4len:
		{
			if _, err := buff.Write([]byte{0}); err != nil {
				logrus.Fatal(err)
			}
		}
	case net.IPv6len:
		{
			if _, err := buff.Write([]byte{1}); err != nil {
				logrus.Fatal(err)
			}
		}
	default:
		logrus.Fatal("invalid netaddr")
	}

	if _, err := buff.Write(h.SenderAddr.IP); err != nil {
		logrus.Fatal(err)
	}

	binary.Write(buff, binary.BigEndian, uint16(h.SenderAddr.Port))

	// Write Recv addr
	switch len(h.ReceiverAddr.IP) {
	case net.IPv4len:
		{
			if _, err := buff.Write([]byte{0}); err != nil {
				logrus.Fatal(err)
			}
		}
	case net.IPv6len:
		{
			if _, err := buff.Write([]byte{1}); err != nil {
				logrus.Fatal(err)
			}
		}
	default:
		logrus.Fatal("invalid netaddr")
	}

	if _, err := buff.Write(h.ReceiverAddr.IP); err != nil {
		logrus.Fatal(err)
	}
	binary.Write(buff, binary.BigEndian, uint16(h.ReceiverAddr.Port))

	// Write user agent [len][string]
	binary.Write(buff, binary.BigEndian, uint64(len(h.UserAgent)))
	buff.WriteString(h.UserAgent)

	return buff.Bytes()
}

func (h *hand) Type() uint8 {
	return consensus.MsgTypeHand
}

func (h *hand) Read(r io.Reader) error {

	if err := binary.Read(r, binary.BigEndian, &h.Version); err != nil {
		return err
	}

	if h.Version != consensus.ProtocolVersion {
		return errors.New("incompatibility protocol version")
	}

	if err := binary.Read(r, binary.BigEndian, (*uint32)(&h.Capabilities)); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &h.Nonce); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, (*uint64)(&h.TotalDifficulty)); err != nil {
		return err
	}

	// read Sender addr
	var ipFlag int8
	var ipAddr []byte
	var ipPort uint16

	if err := binary.Read(r, binary.BigEndian, &ipFlag); err != nil {
		return err
	}

	if ipFlag == 0 {
		// for ipv4 addr
		ipAddr = make([]byte, net.IPv4len)
	} else {
		// for ipv6 addr
		ipAddr = make([]byte, net.IPv6len)
	}

	if _, err := io.ReadFull(r, ipAddr); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &ipPort); err != nil {
		return err
	}

	h.SenderAddr = &net.TCPAddr{
		IP: ipAddr,
		Port: int(ipPort),
	}

	// read Recv addr
	if err := binary.Read(r, binary.BigEndian, &ipFlag); err != nil {
		return err
	}

	if ipFlag == 0 {
		// for ipv4 addr
		ipAddr = make([]byte, net.IPv4len)
	} else {
		// for ipv6 addr
		ipAddr = make([]byte, net.IPv6len)
	}

	if _, err := io.ReadFull(r, ipAddr); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, &ipPort); err != nil {
		return err
	}

	h.ReceiverAddr = &net.TCPAddr{
		IP: ipAddr,
		Port: int(ipPort),
	}

	// read user agent
	var userAgentLen uint64
	if err := binary.Read(r, binary.BigEndian, &userAgentLen); err != nil {
		return err
	}

	logrus.Debug("userAgentlen: ", userAgentLen)

	buff := make([]byte, userAgentLen)
	if _, err := io.ReadFull(r, buff); err != nil {
		return err
	}

	h.UserAgent = string(buff)
	return nil
}

// Second part of a handshake, receiver of the first part replies with its own
// version and characteristics.
type shake struct {
	// protocol version of the sender
	Version uint32
	// capabilities of the sender
	Capabilities consensus.Capabilities
	// total difficulty accumulated by the sender, used to check whether sync
	// may be needed
	TotalDifficulty consensus.Difficulty

	// name of version of the software
	UserAgent string
}

func (h *shake) Bytes() []byte {
	logrus.Info("shake struct to bytes")
	buff := new(bytes.Buffer)

	if err := binary.Write(buff, binary.BigEndian, h.Version); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint32(h.Capabilities)); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(h.TotalDifficulty)); err != nil {
		logrus.Fatal(err)
	}

	// Write user agent [len][string]
	binary.Write(buff, binary.BigEndian, uint64(len(h.UserAgent)))
	buff.WriteString(h.UserAgent)

	return buff.Bytes()
}

func (h *shake) Type() uint8 {
	return consensus.MsgTypeShake
}

func (h *shake) Read(r io.Reader) error {

	if err := binary.Read(r, binary.BigEndian, &h.Version); err != nil {
		return err
	}

	if h.Version != consensus.ProtocolVersion {
		return errors.New("incompatibility protocol version")
	}

	if err := binary.Read(r, binary.BigEndian, (*uint32)(&h.Capabilities)); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, (*uint64)(&h.TotalDifficulty)); err != nil {
		return err
	}

	var userAgentLen uint64
	if err := binary.Read(r, binary.BigEndian, &userAgentLen); err != nil {
		return err
	}

	logrus.Debug("userAgentlen: ", userAgentLen)

	buff := make([]byte, userAgentLen)
	if _, err := io.ReadFull(r, buff); err != nil {
		return err
	}

	h.UserAgent = string(buff)
	return nil
}

// shakeByHand sends hand to receive shake
func shakeByHand(conn net.Conn) (*shake, error) {

	logrus.Info("start peer shakeByHand")
	// create hand
	sender := conn.LocalAddr().(*net.TCPAddr)
	receiver := conn.RemoteAddr().(*net.TCPAddr)
	nonce := uint64(1)

	msg := hand {
		Version:         consensus.ProtocolVersion,
		Capabilities:    consensus.CapFullNode,
		Nonce:           nonce,
		TotalDifficulty: consensus.Difficulty(1),
		SenderAddr:      sender,
		ReceiverAddr:    receiver,
		UserAgent:       userAgent,
	}

	logrus.Info("send hand to peer")
	// Send own hand
	if _, err := WriteMessage(conn, &msg); err != nil {
		return nil, err
	}

	logrus.Info("recv shake from peer")

	// Read peer shake
	// TODO: check nonce
	sh := new(shake)
	if _, err := ReadMessage(conn, sh); err != nil {
		return nil, err
	}
	logrus.Debug("receive shake: ", sh)

	return sh, nil
}

// handByShake sends shake and return received hand
func handByShake(conn net.Conn) (*hand, error) {

	logrus.Info("start peer shakeByHand")
	var h hand
	// Recv remote hand
	if _, err := ReadMessage(conn, &h); err != nil {
		return nil, err
	}

	logrus.Debug("receive shake: ", h)

	// Read peer shake
	// TODO: check nonce
	msg := shake {
		Version: consensus.ProtocolVersion,
		Capabilities: consensus.CapFullNode,
		TotalDifficulty: consensus.Difficulty(1),
		UserAgent: userAgent,

	}
	if _, err := WriteMessage(conn, &msg); err != nil {
		return nil, err
	}


	return &h, nil
}
