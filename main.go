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

var info = log.New(os.Stdout, "INFO: ", log.Ldate | log.Ltime)
var warn = log.New(os.Stderr, "WARN: ", log.Ldate | log.Ltime)
var upstreamAddr, downstreamAddr, downstreamMac string

func main() {
	flag.StringVar(&upstreamAddr, "u", "127.0.0.1:6666", "The IP address and port to listen on")
	flag.StringVar(&downstreamAddr, "d", "10.0.0.1:7777", "The IP address and port to forward to")
	flag.StringVar(&downstreamMac, "m", "aa:bb:cc:dd:ee:ff", "The MAC address to wake up")
	flag.Parse()
	if flag.NArg() > 0 {
		warn.Println("All flags are mandatory and no other arguments are accepted")
		flag.PrintDefaults()
		os.Exit(2);
	}
	srv, err := net.Listen("tcp", upstreamAddr)
	if err != nil {
		panic(err)
	}
	defer srv.Close()
	info.Printf("Listening on %s, Waking %s, Proxing to %s", upstreamAddr, downstreamMac, downstreamAddr)

	for {
		conn, err := srv.Accept()
		if err != nil {
			panic(err)
		}
		info.Println("Upstream connection attempt from", conn.RemoteAddr())
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
			info.Printf("Cannot connect to downstream server on behalf of %s, sending wakeup (attempt %d)", upstream.RemoteAddr(), try+1)
			if packet, err := gowol.NewMagicPacket(downstreamMac); err == nil {
				packet.Send("255.255.255.255") // send to broadcast
			}
			time.Sleep(10 * time.Second)
			try++
			if try > 9 {
				info.Println("Giving up last attempt to connect on behalf of", upstream.RemoteAddr())
				return
			}
		} else {
			break
		}
	}
	if err == nil {
		defer downstream.Close()
	}

	info.Printf("Connection to %s on behalf of %s established", downstreamAddr, upstream.RemoteAddr())

	ctx, cancel := context.WithCancel(context.Background())
	go copySocket(cancel, upstream, downstream)
	go copySocket(cancel, downstream, upstream)

	<-ctx.Done()

	info.Printf("Connection for %s closed", upstream.RemoteAddr())
}
