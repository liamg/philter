package main

import (
	"os"
	"os/signal"

	"github.com/liamg/philter/internal/app/philter/blacklist"
	"github.com/liamg/philter/internal/app/philter/server"
	log "github.com/sirupsen/logrus"
)

func main() {

	log.Infof("Loading blacklist from disk...")
	blacklist, err := blacklist.FromFile("./blacklist.txt")
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	log.Infof("Starting server...")
	s := server.New(blacklist)
	s.Start(53)
	<-c
}
