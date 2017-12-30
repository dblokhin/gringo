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
	Version         uint32
	// capabilities of the sender
	Capabilities    capabilities
	// randomly generated for each handshake, helps detect self
	Nonce           uint64
	// total difficulty accumulated by the sender, used to check whether sync
	// may be needed
	TotalDifficulty consensus.Difficulty
	// network address of the sender
	SenderAddr      *net.TCPAddr
	ReceiverAddr    *net.TCPAddr

	// name of version of the software
	UserAgent       string
}

func (h hand) Bytes() []byte {
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
	case net.IPv4len: {
		if _, err := buff.Write([]byte{0}); err != nil {
			logrus.Fatal(err)
		}
	}
	case net.IPv6len: {
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
	case net.IPv4len: {
		if _, err := buff.Write([]byte{0}); err != nil {
			logrus.Fatal(err)
		}
	}
	case net.IPv6len: {
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

// Second part of a handshake, receiver of the first part replies with its own
// version and characteristics.
type shake struct {
	// protocol version of the sender
	Version         uint32
	// capabilities of the sender
	Capabilities    capabilities
	// total difficulty accumulated by the sender, used to check whether sync
	// may be needed
	TotalDifficulty consensus.Difficulty

	// name of version of the software
	UserAgent       string
}

func (h *shake) Read(r io.Reader) error {

	if err := binary.Read(r, binary.BigEndian, &h.Version); err != nil {
		return err
	}

	if h.Version != protocolVersion {
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

	if userAgentLen > maxUserAgentLength {
		logrus.Warn("too big userAgent len value")
		return errors.New("invalid userAgent len value")
	}

	buff := make([]byte, userAgentLen)
	if _, err := io.ReadFull(r, buff); err != nil {
		return err
	}

	h.UserAgent = string(buff)
	return nil
}

func handshake(conn net.Conn) (*shake, error) {

	logrus.Info("start peer handshake")
	// create handshake
	sender := conn.LocalAddr().(*net.TCPAddr)
	receiver := conn.RemoteAddr().(*net.TCPAddr)
	nonce := uint64(1)

	msg := hand {
		Version: protocolVersion,
		Capabilities: fullNode,
		Nonce: nonce,
		TotalDifficulty: consensus.Difficulty(1),
		SenderAddr: sender,
		ReceiverAddr: receiver,
		UserAgent: userAgent,
	}

	logrus.Info("send hand to peer")
	// Send own hand
	if err := writeMessage(conn, msgTypeHand, msg.Bytes()); err != nil {
		return nil, err
	}

	logrus.Info("recv shake from peer")

	// Read peer shake
	sh := new(shake)
	if err := readMessage(conn, sh); err != nil {
		return nil, err
	}
	logrus.Debug(sh)

	return sh, nil
}
