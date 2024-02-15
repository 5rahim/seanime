package qbittorrent

import (
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClient_Start(t *testing.T) {

	client := NewClient(&NewClientOptions{
		Logger:   util.NewLogger(),
		Username: "admin",
		Password: "adminadmin",
		Port:     8081,
		Host:     "127.0.0.1",
	})

	err := client.Start()
	assert.Nil(t, err)

}

func TestClient_CheckStart(t *testing.T) {

	client := NewClient(&NewClientOptions{
		Logger:   util.NewLogger(),
		Username: "admin",
		Password: "adminadmin",
		Port:     8081,
		Host:     "127.0.0.1",
	})

	started := client.CheckStart()
	assert.True(t, started)

}
