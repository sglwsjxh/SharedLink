package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sharedlink",
	Short: "LAN P2P file transfer tool",
	Long: `SharedLink is a peer-to-peer file transfer tool for local networks.
It uses mDNS for automatic device discovery and TCP for fast file transfer.

Usage:
  sharedlink send <file>       Send a file (listen for incoming connection)
  sharedlink recv              Scan LAN and receive a file
  sharedlink recv <ip>         Receive from a specific address (default port 53349)
  sharedlink recv <ip>:<port>  Receive from a specific address with custom port`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func ExecuteContext(ctx context.Context) {
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
