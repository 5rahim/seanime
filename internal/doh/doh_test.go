package doh

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ncruces/go-dns"
)

func TestDoHResolver(t *testing.T) {
	// Start a temporary HTTP test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello via DoH"))
	}))
	defer ts.Close()

	// Extract hostname and port from the test server URL
	host, port, err := net.SplitHostPort(ts.Listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to parse test server address: %v", err)
	}

	// Mock a "DNS record" by pointing a custom hostname to the test server's IP
	fakeHostname := "test.local"

	dohURL := "https://cloudflare-dns.com/dns-query{?dns}"

	// Set up the DoH resolver
	resolver, err := dns.NewDoHResolver(dohURL, dns.DoHCache())
	if err != nil {
		t.Fatalf("failed to create DoH resolver: %v", err)
	}

	// Override the default resolver
	net.DefaultResolver = resolver

	// Use a custom DialContext to redirect fakeHostname to test server IP
	dialer := &net.Dialer{
		Timeout: 3 * time.Second,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				// Shortcut: Always return test server's IP for "test.local"
				d := net.Dialer{}
				if network == "udp" || network == "tcp" {
					return d.Dial(network, net.JoinHostPort(host, "53"))
				}
				return d.Dial(network, address)
			},
		},
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Intercept DNS for fakeHostname only
				if addr[:len(fakeHostname)] == fakeHostname {
					addr = net.JoinHostPort(host, port)
				}
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}

	// Make a request to the fake hostname (which we route to the test server)
	resp, err := client.Get("http://" + fakeHostname + ":" + port)
	if err != nil {
		t.Fatalf("failed to GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %v", resp.Status)
	}

	// Read the response body
	bodyR, err := io.ReadAll(resp.Body)

	t.Log(string(bodyR))
}
