package tap

import (
	"net"

	gophertun "github.com/m13253/gophertun"
)

func parseCIDR(cidr string) *net.IPNet {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	return &net.IPNet{
		IP:   ip,
		Mask: ipnet.Mask,
	}
}

// data := gophertun.Packet{}
// json.Unmarshal(nil, &data)
// go func() {
// 	for range time.Tick(5 * time.Second) {
// 		fmt.Println(t.Write(&data, true))
// 	}
// }()

func NewTapNet(ip string, callback func(packet *gophertun.Packet)) (gophertun.Tunnel, error) {
	c := &gophertun.TunTapConfig{
		NameHint:              "OwO",
		AllowNameSuffix:       true,
		PreferredNativeFormat: gophertun.FormatEthernet,
	}

	t, e := c.Create()
	if e != nil {
		return nil, e
	}
	//defer t.Close()

	e = t.SetMTU(65521)
	if e != nil {
		return nil, e
	}

	_, e = t.AddIPAddresses([]*gophertun.IPAddress{
		{
			Net:  parseCIDR(ip + "/24"),
			Peer: parseCIDR("10.42.0.1/24"),
		},
	})
	if e != nil {
		return nil, e
	}

	e = t.Open(gophertun.FormatEthernet)
	if e != nil {
		return nil, e
	}

	if callback != nil {
		go func() {
			for {
				p, e := t.Read()
				if e != nil || p == nil {
					continue
				}

				callback(p)
			}
		}()
	}
	return t, nil
}
