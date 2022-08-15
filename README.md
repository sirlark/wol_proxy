# wol_proxy

A bare-bones pass-through proxy that will send wake-on-lan when it can't
connect to the server

`wol_proxy` can be run to listen on a single command line specified port and
proxy requests received to another specifed IP address and port. When a request
is received, `wol_proxy` will attempt to connected to the downstream server
specified adress and port. If this fails, a wake-on-lan magic packet for a
command line specifed MAC address is broadcast, and the request is retried
every 10 seconds for up to 90 seconds.


## Usage

```
Usage of ./wol_proxy:
  -d string
    	The IP address and port to forward to (default "10.0.0.1:7777")
  -m string
    	The MAC address to wake up (default "aa:bb:cc:dd:ee:ff")
  -u string
    	The IP address and port to listen on (default "127.0.0.1:6666")
```
