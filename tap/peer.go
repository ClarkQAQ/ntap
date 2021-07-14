package tap

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func NewPeer(remoteAddr string, remotePort int, localPort int, callback func(msg []byte)) (*net.UDPConn, error) {
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: localPort}
	dstAddr := &net.UDPAddr{IP: net.ParseIP(remoteAddr), Port: remotePort}

	sdtConn, e := net.DialUDP("udp", srcAddr, dstAddr)
	if e != nil {
		return nil, e
	}
	

	if _, e = sdtConn.Write([]byte("hello, I'm new peer:" + fmt.Sprint(net.IPv4zero))); e != nil {
		return nil, e
	}

	data := make([]byte, 1024)
	n, remoteUDPAddr, e := sdtConn.ReadFromUDP(data)
	if e != nil {
		return nil, e
	}

	sdtConn.Close()

	anotherPeer := parseAddr(string(data[:n]))
	fmt.Printf("local:%s server:%s another:%s\n", srcAddr, remoteUDPAddr, anotherPeer.String())

	pwpConn, e := net.DialUDP("udp", srcAddr, &anotherPeer)

	if e == nil { // 打开连接 && 保活
		go func() {
			for c := range time.NewTicker(5 * time.Second).C {
				if _, e := pwpConn.Write([]byte(fmt.Sprintf("%v", c))); e != nil {
					break
				}
			}
		}()
	}

	if callback != nil && e == nil {
		go func() {
			for {
				data := make([]byte, 65521)
				n, _, e := pwpConn.ReadFromUDP(data)
				if e == nil && data[:n] != nil {
					callback(data[:n])
				}
			}
		}()
	}

	return pwpConn, e
}

func parseAddr(addr string) net.UDPAddr {
	t := strings.Split(addr, ":")
	port, _ := strconv.Atoi(t[1])
	return net.UDPAddr{
		IP:   net.ParseIP(t[0]),
		Port: port,
	}
}
