package doh

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/likexian/doh-go"
	"github.com/likexian/doh-go/dns"
	"github.com/stretchr/testify/assert"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestDoH(t *testing.T) {

	// init a context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := doh.Use(doh.CloudflareProvider)

	// do doh query
	rsp, err := c.Query(ctx, "nyaa.si", dns.TypeA)
	if err != nil {
		panic(err)
	}

	// close the client
	c.Close()

	// doh dns answer
	answer := rsp.Answer

	ip := ""
	// print all answer
	for _, a := range answer {
		if ip == "" {
			ip = a.Data
		}
		fmt.Printf("%s -> %s\n", a.Name, a.Data)
	}

	fmt.Println("IP: ", ip)

	assert.NotEmpty(t, ip)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("GET", "https://"+ip, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	buff, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(buff))

	assert.Equal(t, 200, resp.StatusCode)

}

func TestDoH2(t *testing.T) {

	req, err := http.NewRequest("GET", "https://nyaa.si", nil)
	if err != nil {
		t.Fatal(err)
	}

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: 5 * time.Second,
				}
				return d.DialContext(ctx, "udp", "1.1.1.1:53")
			},
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	buff, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(buff))

	assert.Equal(t, 200, resp.StatusCode)

}
