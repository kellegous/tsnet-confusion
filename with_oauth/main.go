package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2/clientcredentials"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tsnet"
)

func getAuthKey(ctx context.Context) (string, error) {
	cfg := clientcredentials.Config{
		ClientID:     os.Getenv("TS_CLIENT_ID"),
		ClientSecret: os.Getenv("TS_CLIENT_SECRET"),
		TokenURL:     "https://api.tailscale.com/api/v2/oauth/token",
	}

	t, err := cfg.Token(ctx)
	if err != nil {
		return "", err
	}

	c := tailscale.NewClient("-", tailscale.APIKey(t.AccessToken))

	key, _, err := c.CreateKey(
		ctx,
		tailscale.KeyCapabilities{
			Devices: tailscale.KeyDeviceCapabilities{
				Create: tailscale.KeyDeviceCreateCapabilities{
					Reusable: true,
					Tags:     []string{"tag:service"},
				},
			},
		},
	)

	return key, err
}

func main() {
	tailscale.I_Acknowledge_This_API_Is_Unstable = true
	ctx := context.Background()
	key, err := getAuthKey(ctx)
	if err != nil {
		log.Panic(err)
	}

	os.MkdirAll("data/oauth", 0700)
	s := tsnet.Server{
		Dir:      "data/oauth",
		Hostname: "service",
		AuthKey:  key,
		Logf: func(f string, args ...any) {
			// remove to see logs
		},
	}
	defer s.Close()

	c, err := s.LocalClient()
	if err != nil {
		log.Panic(err)
	}

	var status *ipnstate.Status
	for {
		status, err = c.Status(ctx)
		if err != nil {
			log.Panic(err)
		}

		if status.BackendState == "Running" {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
	h, _, _ := strings.Cut(status.Self.DNSName, ".")
	fmt.Println(h)
}
