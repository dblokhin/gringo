package p2p

import (
	"io"
	"bufio"
	"github.com/sirupsen/logrus"
	"errors"
	"net"
)

// SendableMessage defines methods for WriteMessage function
type SendableMessage interface {
	// Bytes returns binary data of body message
	Bytes() []byte
	// Type says whats the message type should use in header
	Type() uint8
}

// ReadableMessage defines methods for ReadMessage function
type ReadableMessage interface {
	Read(r io.Reader) error

	//expected type of receiving message
	Type() uint8
}

// WriteMessage writes to wr (net.conn) protocol message
func WriteMessage(w io.Writer, msg SendableMessage) error {
	data := msg.Bytes()

	header := msgHeader{
		magic:   magicCode,
		msgType: msg.Type(),
		msgLen:  uint64(len(data)),
	}

	// use the buffered writer
	wr := bufio.NewWriter(w)
	if err := header.Write(wr); err != nil {
		return err
	}

	if _, err := wr.Write(data); err != nil {
		return err
	}

	return wr.Flush()
}

// ReadMessage reads from r (net.conn) protocol message
func ReadMessage(r io.Reader, msg ReadableMessage) error {
	var header msgHeader

	// get the msg header
	rh := io.LimitReader(r, headerLen)
	if err := header.Read(rh); err != nil {
		return err
	}
	logrus.Debug("readed header: ", header)

	if header.msgType != msg.Type() {
		return errors.New("receive unexpected message type")
	}

	rb := io.LimitReader(r, int64(header.msgLen))
	return msg.Read(rb)
}

// HandleLoop starts event loop listening
func HandleLoop(conn net.Conn) error {
	input := bufio.NewReader(conn)

	for {
		header := new(msgHeader)
		if err := header.Read(input); err != nil {
			return err
		}
		logrus.Debug("receive header: ", header)

		if header.msgLen > maxMessageSize {
			return errors.New("too big message size")
		}

		switch header.msgType {
		case msgTypePing: // send pong
		case msgTypePong: // nothing
		case msgTypeGetPeerAddrs:
		case msgTypePeerAddrs:
		case msgTypeGetHeaders:
		case msgTypeHeaders:
		case msgTypeGetBlock:
		case msgTypeBlock:
		case msgTypeTransaction:

		default:
			return errors.New("receive unexpected message (type) from peer")
		}
	}
}