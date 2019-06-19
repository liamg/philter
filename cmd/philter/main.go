package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/liamg/philter/internal/app/philter/blacklist"
	"github.com/liamg/philter/internal/app/philter/server"
	"github.com/liamg/philter/internal/app/philter/update"
	"github.com/liamg/philter/internal/app/philter/version"
	log "github.com/sirupsen/logrus"
)

func main() {

	log.Infof("philter [%s] ", version.Version)

	log.Infof("Loading blacklist from disk...")
	blacklist, err := blacklist.FromFile("/var/lib/philter/blacklist.txt")
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	log.Infof("Starting server...")
	s := server.New(blacklist)

	go func() {
		ticker := time.NewTicker(time.Minute * 1)
		defer ticker.Stop()
		for {
			<-ticker.C
			done, err := update.Update()
			if err != nil {
				log.Errorf("update failed: %s", err)
				continue
			}
			if done {
				log.Infof("update applied - restarting...")
				close(c)
				return
			}
		}
	}()

	if err := s.Start(53); err != nil {
		panic(err)
	}
	<-c
}
