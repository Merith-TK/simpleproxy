package main

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
)

func handleTCPConnection(src net.Conn, targetAddr string) {
	defer src.Close()

	dst, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("[ERROR] Unable to connect to target: %v\n", err)
		return
	}
	defer dst.Close()

	done := make(chan struct{})

	go func() {
		io.Copy(dst, src)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(src, dst)
		done <- struct{}{}
	}()

	<-done
}

func startTCPProxy(ctx context.Context, wg *sync.WaitGroup, localAddr, targetAddr string) {
	defer wg.Done()

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("[ERROR] Unable to listen on %s: %v\n", localAddr, err)
	}
	defer listener.Close()

	log.Printf("[INFO] Listening on %s (TCP), forwarding to %s\n", localAddr, targetAddr)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Shutting down TCP proxy on %s\n", localAddr)
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("[ERROR] Failed to accept connection: %v\n", err)
				continue
			}
			go handleTCPConnection(conn, targetAddr)
		}
	}
}
