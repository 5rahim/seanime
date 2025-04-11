package doh

import (
	"context"
	"net"
	"seanime/internal/util"

	"github.com/ncruces/go-dns"
	"github.com/rs/zerolog"
)

func HandleDoH(dohUrl string, logger *zerolog.Logger) {
	defer util.HandlePanicInModuleThen("doh/HandleDoH", func() {})

	if dohUrl == "" {
		return
	}

	logger.Info().Msgf("doh: Using DoH resolver: %s", dohUrl)

	// Set up the DoH resolver
	resolver, err := dns.NewDoHResolver(dohUrl, dns.DoHCache())
	if err != nil {
		logger.Error().Err(err).Msgf("doh: Failed to create DoH resolver: %s", dohUrl)
		return
	}

	// Override the default resolver
	net.DefaultResolver = resolver

	// Test the resolver
	_, err = resolver.LookupIPAddr(context.Background(), "ipv4.google.com")
	if err != nil {
		logger.Error().Err(err).Msgf("doh: DoH resolver failed lookup: %s", dohUrl)
		return
	}
}
