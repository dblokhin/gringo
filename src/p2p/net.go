package p2p

import (
	"io"
	"bufio"
	"github.com/sirupsen/logrus"
	"errors"
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
