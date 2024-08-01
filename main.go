package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

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
		log.Printf("Unable to connect to target: %v\n", err)
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

func startTCPProxy(localAddr, targetAddr string) {
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("Unable to listen on %s: %v\n", localAddr, err)
	}
	defer listener.Close()

	log.Printf("Listening on %s (TCP), forwarding to %s\n", localAddr, targetAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		go handleTCPConnection(conn, targetAddr)
	}
}

func startUDPProxy(localAddr, targetAddr string) {
	localConn, err := net.ListenPacket("udp", localAddr)
	if err != nil {
		log.Fatalf("Unable to listen on %s: %v\n", localAddr, err)
	}
	defer localConn.Close()

	remoteAddr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		log.Fatalf("Unable to resolve target address: %v\n", err)
	}

	buf := make([]byte, 4096)

	log.Printf("Listening on %s (UDP), forwarding to %s\n", localAddr, targetAddr)

	for {
		n, addr, err := localConn.ReadFrom(buf)
		if err != nil {
			log.Printf("Failed to read from connection: %v\n", err)
			continue
		}

		go func(data []byte, addr net.Addr) {
			remoteConn, err := net.DialUDP("udp", nil, remoteAddr)
			if err != nil {
				log.Printf("Unable to connect to target: %v\n", err)
				return
			}
			defer remoteConn.Close()

			_, err = remoteConn.Write(data)
			if err != nil {
				log.Printf("Failed to write to target: %v\n", err)
				return
			}

			n, _, err := remoteConn.ReadFrom(data)
			if err != nil {
				log.Printf("Failed to read from target: %v\n", err)
				return
			}

			_, err = localConn.WriteTo(data[:n], addr)
			if err != nil {
				log.Printf("Failed to write back to source: %v\n", err)
			}
		}(buf[:n], addr)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <config-file>", os.Args[0])
	}

	configFile := os.Args[1]

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v\n", err)
	}

	var config ProxyConfig
	if err := json5.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config file: %v\n", err)
	}

	for _, proxy := range config.Proxy {
		// Default type to "both" if not provided
		if proxy.Type == "" {
			proxy.Type = "both"
		}

		// Default local to the port of remote if not provided
		if proxy.Local == "" {
			parts := strings.Split(proxy.Remote, ":")
			if len(parts) == 2 {
				proxy.Local = ":" + parts[1]
			} else {
				log.Fatalf("Invalid remote address format: %s\n", proxy.Remote)
			}
		}

		switch proxy.Type {
		case "tcp":
			go startTCPProxy(proxy.Local, proxy.Remote)
		case "udp":
			go startUDPProxy(proxy.Local, proxy.Remote)
		case "both":
			go startTCPProxy(proxy.Local, proxy.Remote)
			go startUDPProxy(proxy.Local, proxy.Remote)
		default:
			log.Printf("Unknown proxy type: %s\n", proxy.Type)
		}
	}

	select {}
}
