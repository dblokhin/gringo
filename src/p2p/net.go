package p2p

import (
	"io"
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"bufio"
	"os"
)

// writeMessage writes to wr (net.conn) protocol message
func writeMessage(w io.Writer, mType uint8, body []byte) error {
	logrus.Debug("write message to stream:")
	header := msgHeader{
		magic:   magicCode,
		msgType: mType,
		msgLen:  int64(len(body)),
	}

	// duplicate
	f, _ := os.Create("log")
	defer f.Close()
	w = io.MultiWriter(w, f)


	// use the buffered writer
	wr := bufio.NewWriter(w)
	logrus.Debug("write header to stream")

	if err := header.Write(wr); err != nil {
		return err
	}

	logrus.Debug("write body to stream")
	if _, err := wr.Write(body); err != nil {
		return err
	}

	return wr.Flush()
}

// readMessage reads from r (net.conn) protocol message
func readMessage(r io.Reader, data interface{}) error {
	var header msgHeader

	rh := io.LimitReader(r, headerLen)
	if err := binary.Read(rh, binary.BigEndian, header); err != nil {
		return err
	}

	// TODO: what if data capacity is less than header.msgLen
	rb := io.LimitReader(r, header.msgLen)
	return binary.Read(rb, binary.BigEndian, data)
}