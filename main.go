package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	host       = flag.String("host", "0.0.0.0", "host")
	port       = flag.String("port", "8000", "port")
	buffer     = flag.Int64("buffer", 60, "buffer for recording")
	output     = flag.String("output", "output", "output")
	title      = flag.String("title", "radcast", "title")
	configPath = flag.String("config", "config.json", "path of config.json")
	setup      = flag.Bool("setup", false, "initialize json configuration")
)

func main() {
	flag.Parse()

	if *setup {
		runSetup()
		return
	}

	if err := runRadcast(); err != nil {
		log.Fatal(err)
	}
}

func runRadcast() error {

	converter, err := lookConverterCommand()
	if err != nil {
		return err
	}

	r := NewRadcast(*configPath, *host, *port, *title, *output, *buffer, converter)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGHUP)

	go func() {
		for {
			s := <-signalChan
			r.Log("got signal:", s)
			switch s {
			case syscall.SIGHUP:
				r.ReloadConfig()
			default:
				r.Stop()
			}
		}
	}()

	return r.Run()
}

func runSetup() {
	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	go func() {
		s := <-signalChan
		log.Println("got signal:", s)
		cancel()
	}()

	if err := SetupConfig(ctx); err != nil {
		log.Fatal(err)
	}

}
