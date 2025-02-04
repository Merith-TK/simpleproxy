package main

import (
	"context"
	"log"
	"net"
	"sync"
)

func startUDPProxy(ctx context.Context, wg *sync.WaitGroup, localAddr, targetAddr string) {
	defer wg.Done()

	localConn, err := net.ListenPacket("udp", localAddr)
	if err != nil {
		log.Fatalf("[ERROR] Unable to listen on %s: %v\n", localAddr, err)
	}
	defer localConn.Close()

	remoteAddr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		log.Fatalf("[ERROR] Unable to resolve target address: %v\n", err)
	}

	buf := make([]byte, 4096)

	log.Printf("[INFO] Listening on %s (UDP), forwarding to %s\n", localAddr, targetAddr)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Shutting down UDP proxy on %s\n", localAddr)
			return
		default:
			n, addr, err := localConn.ReadFrom(buf)
			if err != nil {
				log.Printf("[ERROR] Failed to read from connection: %v\n", err)
				continue
			}

			go func(data []byte, addr net.Addr) {
				remoteConn, err := net.DialUDP("udp", nil, remoteAddr)
				if err != nil {
					log.Printf("[ERROR] Unable to connect to target: %v\n", err)
					return
				}
				defer remoteConn.Close()

				_, err = remoteConn.Write(data)
				if err != nil {
					log.Printf("[ERROR] Failed to write to target: %v\n", err)
					return
				}

				n, _, err := remoteConn.ReadFrom(data)
				if err != nil {
					log.Printf("[ERROR] Failed to read from target: %v\n", err)
					return
				}

				_, err = localConn.WriteTo(data[:n], addr)
				if err != nil {
					log.Printf("[ERROR] Failed to write back to source: %v\n", err)
				}
			}(buf[:n], addr)
		}
	}
}
