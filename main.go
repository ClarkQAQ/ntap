package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"ntap/tap"
	"os"

	gophertun "github.com/m13253/gophertun"
)

var callbackPacket = func(packet *gophertun.Packet) {
	fmt.Println(packet)
}

func main() {
	// 虚拟局域网IP
	var vlanIp string
	flag.StringVar(&vlanIp, "ip", "10.42.0.2", "vlan ip address")

	// 远程打洞服务器地址
	var remoteAddr string
	flag.StringVar(&remoteAddr, "raddr", "127.0.0.1", "vlan ip address")

	// 远程打洞服务器端口
	var remotePort int
	flag.IntVar(&remotePort, "rport", 9871, "vlan ip address")

	// 本地端口
	var localPort int
	flag.IntVar(&localPort, "lport", 9872, "vlan ip address")

	// 服务器端口
	var serverPort int
	flag.IntVar(&serverPort, "sport", 9871, "vlan ip address")

	// 节点模式
	var isPeer bool
	flag.BoolVar(&isPeer, "peer", false, "is peer")

	// 服务器模式
	var isServer bool
	flag.BoolVar(&isServer, "server", false, "is server")

	flag.Parse()

	if isServer {
		NewServer(serverPort)
	}

	if isPeer {
		NewPeer(vlanIp, remoteAddr, remotePort, localPort)
	}

	if !isServer && !isPeer {
		fmt.Println("tag: -server or -peer")
	} else {
		select {}
	}
}

func NewPeer(vlanIp, remoteAddr string, remotePort, localPort int) {
	t, e := tap.NewTapNet(vlanIp, func(packet *gophertun.Packet) {
		callbackPacket(packet)
	})

	if e != nil {
		log.Printf("new peer tap net: %s", e)
		os.Exit(1)
	}

	//defer t.Close()

	conn, e := tap.NewPeer(remoteAddr, remotePort, localPort, func(b []byte) {
		data := gophertun.Packet{}
		if json.Unmarshal(b, &data) == nil {
			t.Write(&data, true)
		}
	})

	if e != nil {
		log.Printf("new peer: %s", e)
		os.Exit(1)
	}

	//defer conn.Close()

	callbackPacket = func(packet *gophertun.Packet) {
		if b, e := json.Marshal(packet); e == nil {
			conn.Write(b)
		}
	}
}

func NewServer(serverPort int) {
	listener, e := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: serverPort})
	if e != nil {
		log.Printf("listen server: %s", e)
		os.Exit(1)
	}

	log.Printf("listen: <%s:%d> \n", net.IPv4zero, serverPort)

	peers := make([]net.UDPAddr, 0, 2)
	data := make([]byte, 1024)
	go func() {
		for {
			n, remoteAddr, err := listener.ReadFromUDP(data)
			if err != nil {
				log.Printf("error during read: %s", e)
			}
			log.Printf("<%s> %s\n", remoteAddr.String(), data[:n])
			peers = append(peers, *remoteAddr)
			if len(peers) == 2 {
				log.Printf("%s <--> %s\n", peers[0].String(), peers[1].String())
				listener.WriteToUDP([]byte(peers[1].String()), &peers[0])
				listener.WriteToUDP([]byte(peers[0].String()), &peers[1])
			}
		}
	}()
}
