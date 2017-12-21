package corduroy

import (
	"strconv"
	"os"
	"net"
)

func buildLocalUri(port int) string {
	host := getLocalhost()
	return buildUri(host, port)
}

func buildUri(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}

func getLocalhost() string {
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	return host
}

func getLocalAddresses() ([]*net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	addresses := make([]*net.IP, 0)
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				addresses = append(addresses, &v.IP)
			case *net.IPAddr:
				addresses = append(addresses, &v.IP)
			}
		}
	}
	return addresses, nil
}