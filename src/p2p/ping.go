package p2p

import (
	"consensus"
	"encoding/binary"
	"bytes"
	"github.com/sirupsen/logrus"
	"io"
)

// ping request
type ping struct {
	// total difficulty accumulated by the sender, used to check whether sync
	// may be needed
	TotalDifficulty consensus.Difficulty
	// total height
	Height uint64
}

func (p ping) Bytes() []byte {
	logrus.Info("ping struct to bytes")
	buff := new(bytes.Buffer)

	if err := binary.Write(buff, binary.BigEndian, uint64(p.TotalDifficulty)); err != nil {
		logrus.Fatal(err)
	}

	if err := binary.Write(buff, binary.BigEndian, uint64(p.Height)); err != nil {
		logrus.Fatal(err)
	}

	return buff.Bytes()
}

func (p ping) Type() uint8 {
	return msgTypePing
}

func (p *ping) Read(r io.Reader) error {

	if err := binary.Read(r, binary.BigEndian, (*uint64)(&p.TotalDifficulty)); err != nil {
		return err
	}

	if err := binary.Read(r, binary.BigEndian, (*uint64)(&p.Height)); err != nil {
		return err
	}

	return nil
}

// pong response same as ping
type pong struct {
	ping
}

func (p pong) Type() uint8 {
	return msgTypePong
}
