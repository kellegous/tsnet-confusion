package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"tailscale.com/client/tailscale"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tsnet"
)

func main() {
	tailscale.I_Acknowledge_This_API_Is_Unstable = true
	ctx := context.Background()

	os.MkdirAll("data/authkey", 0700)
	s := tsnet.Server{
		Dir:      "data/authkey",
		Hostname: "service",
		AuthKey:  os.Getenv("TS_AUTHKEY"),
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
