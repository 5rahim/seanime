package doh

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/likexian/doh-go"
	"github.com/likexian/doh-go/dns"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestDoH(t *testing.T) {

	// init a context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// init doh client, auto select the fastest provider base on your like
	// you can also use as: c := doh.Use(), it will select from all providers
	c := doh.Use(doh.CloudflareProvider)

	// do doh query
	rsp, err := c.Query(ctx, "animetosho.org", dns.TypeA)
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

	req, err := http.NewRequest("GET", "http://"+ip, nil)
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
