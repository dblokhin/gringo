package p2p

import (
	"io"
	"bufio"
	"github.com/sirupsen/logrus"
	"errors"
)

// writeMessage writes to wr (net.conn) protocol message
func writeMessage(w io.Writer, mType uint8, body []byte) error {
	header := msgHeader{
		magic:   magicCode,
		msgType: mType,
		msgLen:  uint64(len(body)),
	}

	// use the buffered writer
	wr := bufio.NewWriter(w)
	if err := header.Write(wr); err != nil {
		return err
	}

	if _, err := wr.Write(body); err != nil {
		return err
	}

	return wr.Flush()
}

// readMessage reads from r (net.conn) protocol message
func readMessage(r io.Reader, expectedType uint8, body structReader) error {
	var header msgHeader

	// get the msg header
	rh := io.LimitReader(r, headerLen)
	if err := header.Read(rh); err != nil {
		return err
	}
	logrus.Debug("readed header: ", header)

	if header.msgType != expectedType {
		return errors.New("receive unexpected message type")
	}

	rb := io.LimitReader(r, int64(header.msgLen))
	return body.Read(rb)
}