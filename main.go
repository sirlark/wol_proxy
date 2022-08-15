package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/linde12/gowol"
)

const (
	PROXY_ADDR      = "0.0.0.0:8123"
	SERVER_ADDR     = "10.0.0.1:8096"
	SERVER_MAC_ADDR = "aa:bb:cc:dd:ee:ff"
)

func main() {
	fmt.Println("Hello sockets")

	srv, err := net.Listen("tcp", PROXY_ADDR)
	if err != nil {
		panic(err)
	}
	defer srv.Close()
	fmt.Println("Listening on", PROXY_ADDR)

	for {
		conn, err := srv.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Println("Got client from", conn.RemoteAddr())
		go processConn(conn)
	}
}

func copySocket(cancel context.CancelFunc, dst, src net.Conn) {
	io.Copy(dst, src)
	cancel()
}

func processConn(client net.Conn) {
	defer client.Close()

	var server net.Conn
	var err error
	var try int
	for {
		server, err = net.Dial("tcp", SERVER_ADDR)
		if err != nil {
			fmt.Println("Failed to connect to client")
			if packet, err := gowol.NewMagicPacket(SERVER_MAC_ADDR); err == nil {
				packet.Send("255.255.255.255") // send to broadcast
			}
			time.Sleep(70 * time.Second)
			try++
			if try > 3 {
				fmt.Println("Giving Up")
				return
			}
		} else {
			break
		}
	}
	defer server.Close()

	fmt.Println("Connected socket to:", SERVER_ADDR)

	ctx, cancel := context.WithCancel(context.Background())
	// io.Copy(dst, src)
	go copySocket(cancel, client, server)
	go copySocket(cancel, server, client)

	<-ctx.Done()

	fmt.Println("Clients disconnected")
}
