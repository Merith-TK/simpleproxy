package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/yosuke-furukawa/json5/encoding/json5"
)

type ProxyConfig struct {
	Proxy []Proxy `json:"proxy"`
}

type Proxy struct {
	Local  string `json:"local,omitempty"`
	Remote string `json:"remote"`
	Type   string `json:"type,omitempty"` // "tcp", "udp", or "both"
}

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

func main() {
	flag.Parse()
	configPath := os.Getenv("GOPROXY_CONFIG")
	if configPath == "" || flag.Arg(0) == "" {
		configPath = "goproxy.json" // Default path for Docker
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to read config file (%s): %v\n", configPath, err)
	}

	var config ProxyConfig
	if err := json5.Unmarshal(configData, &config); err != nil {
		log.Fatalf("[ERROR] Failed to parse config file: %v\n", err)
	}

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
