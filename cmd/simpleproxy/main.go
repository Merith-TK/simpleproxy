package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func main() {
	flag.Parse()
	configPath := os.Getenv("GOPROXY_CONFIG")
	if configPath == "" || flag.Arg(0) == "" {
		configPath = "goproxy.json" // Default path for Docker
	}

	config := ReadConfig(configPath)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	for _, proxy := range config.Proxy {
		if proxy.Type == "" {
			proxy.Type = "both"
		}

		if proxy.Local == "" {
			parts := strings.Split(proxy.Remote, ":")
			if len(parts) == 2 {
				proxy.Local = ":" + parts[1]
			} else {
				log.Fatalf("[ERROR] Invalid remote address format: %s\n", proxy.Remote)
			}
		}

		wg.Add(1)
		switch proxy.Type {
		case "tcp":
			go startTCPProxy(ctx, &wg, proxy.Local, proxy.Remote)
		case "udp":
			go startUDPProxy(ctx, &wg, proxy.Local, proxy.Remote)
		case "both":
			go startTCPProxy(ctx, &wg, proxy.Local, proxy.Remote)
			go startUDPProxy(ctx, &wg, proxy.Local, proxy.Remote)
		default:
			log.Printf("[WARNING] Unknown proxy type: %s\n", proxy.Type)
			wg.Done()
		}
	}

	// Handle termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("[INFO] Shutting down proxy server...")
	cancel()
	wg.Wait()
	log.Println("[INFO] Proxy server stopped.")
}
