package main

import (
	"flag"
	"github.com/ghosts-network/news-feed/app/api"
	"github.com/ghosts-network/news-feed/app/listener"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	serverEnabled := flag.Bool("server.enable", false, "Enable/disable web server")
	listenedEnabled := flag.Bool("listener.enable", false, "Enable/disable events listener")

	flag.Parse()

	sigc := make(chan os.Signal, 1)
	lsigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(lsigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	if *serverEnabled {
		go api.RunServer()
	}

	if *listenedEnabled {
		go listener.NewListener(lsigc).Run()
	}

	<-sigc
}
