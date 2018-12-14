package main

import (
	"github.com/dblokhin/gringo/p2p"
	"github.com/sirupsen/logrus"
	"github.com/dblokhin/gringo/chain"
	"github.com/dblokhin/gringo/storage"
	"os"
)

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	logrus.Info("Starting")
	chain := chain.New(&chain.Testnet1, storage.NewSqlStorage(nil))

	sync := p2p.NewSyncer([]string{"127.0.0.1:13414"}, chain, nil)
	sync.Pool.Run()
}

