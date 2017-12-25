package corduroy

import (
	"net"
	"os"
	"strconv"
	"bytes"
	"io/ioutil"
	"net/http"
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

func send(verb string, uri string, body string, visited []int, hops int) (int, string, error) {
	b1 := []byte(body)
	buff := bytes.NewBuffer(b1[:])
	request, err := http.NewRequest(verb, uri, buff)
	if err != nil {
		return 0, "", err
	}

	v := ""
	for _, id := range visited {
		if v != "" {
			v = v + ","
		}
		v = v + strconv.Itoa(id)
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set(visitedHeader, v)
	request.Header.Set(hopsHeader, strconv.Itoa(hops))
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return 0, "", err
	}

	defer response.Body.Close()
	b2, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, "", err
	}

	return response.StatusCode, string(b2), nil
}