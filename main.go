package main

import (
	"context"
	"log"
	"io"
	"os"
	"flag"
	"net"
	"time"

	"github.com/linde12/gowol"
)

var upstreamAddr, downstreamAddr, downstreamMac string

func main() {
	flag.StringVar(&upstreamAddr, "u", "127.0.0.1:6666", "The IP address and port to listen on")
	flag.StringVar(&downstreamAddr, "d", "10.0.0.1:7777", "The IP address and port to forward to")
	flag.StringVar(&downstreamMac, "m", "aa:bb:cc:dd:ee:ff", "The MAC address to wake up")
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Println("All flags are mandatory and no other arguments are accepted")
		flag.PrintDefaults()
		os.Exit(2);
	}
	srv, err := net.Listen("tcp", upstreamAddr)
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

func processConn(upstream net.Conn) {
	defer upstream.Close()

	var downstream net.Conn
	var err error
	var try int
	for {
		downstream, err = net.Dial("tcp", downstreamAddr)
		if err != nil {
			fmt.Println("Failed to connect to client")
			if packet, err := gowol.NewMagicPacket(downstreamMac); err == nil {
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
	go copySocket(cancel, upstream, downstream)
	go copySocket(cancel, downstream, upstream)

	<-ctx.Done()

	fmt.Println("Clients disconnected")
}
