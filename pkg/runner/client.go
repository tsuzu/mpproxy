package runner

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/cs3238-tsuzu/multipath-proxy/pkg/client"
	"github.com/cs3238-tsuzu/multipath-proxy/pkg/config"
)

// runClient runs client-side listener
func runClient(cfg *config.Config) error {
	if cfg.Mode != config.ModeClient {
		return fmt.Errorf("config mode should be client")
	}

	listener, err := net.Listen("tcp", cfg.Client.Endpoint)

	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", cfg.Client.Endpoint, err)
	}

	client, err := client.NewClient(context.Background(), cfg.Client.Peers)

	if err != nil {
		return fmt.Errorf("failed to initialize multipath client: %w", err)
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			if err, ok := err.(net.Error); ok {
				if !err.Temporary() {
					return fmt.Errorf("Accept returned a critical error: %w", err)
				}
			}

			return fmt.Errorf("failed to accept new connection: %w", err)
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			mpconn, err := client.NewMultipathConn(ctx)

			if err != nil {
				log.Printf("failed to prepare multipath connection: %+v", err)

				return
			}

			go io.Copy(conn, mpconn)
			go io.Copy(mpconn, conn)
		}()
	}
}