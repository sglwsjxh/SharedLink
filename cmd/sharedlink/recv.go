package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"

	"SharedLink/internal/transfer"
	"SharedLink/internal/ui"
)

var recvCmd = &cobra.Command{
	Use:   "recv [ip:port]",
	Short: "Receive a file from LAN",
	Long: `Receive a file from a sender on the local network.

Without arguments, scans the LAN for available senders via mDNS.
With an IP:port argument, connects directly to that address.

Examples:
  sharedlink recv              # Scan LAN and pick a sender
  sharedlink recv 192.168.1.100:53349  # Connect directly`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var addr string

		if len(args) == 1 {
			addr = args[0]
			if !strings.Contains(addr, ":") {
				return fmt.Errorf("invalid address %q: expected ip:port format", addr)
			}
		} else {
			scanModel := ui.NewScanModel()
			scanProgram := tea.NewProgram(scanModel)
			result, err := scanProgram.Run()
			if err != nil {
				return err
			}
			scanModel = result.(ui.ScanModel)
			if scanModel.SelectedAddr() == "" {
				return fmt.Errorf("no sender selected")
			}
			addr = scanModel.SelectedAddr()
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		model := ui.NewRecvModel()
		program := tea.NewProgram(model)

		go func() {
			err := transfer.Receive(ctx, addr, func(received int64, total int64) {
				program.Send(ui.RecvProgressMsg{
					ReceivedBytes: received,
					TotalBytes:    total,
				})
			})
			if err != nil {
				program.Send(ui.RecvProgressMsg{Error: err})
			} else {
				program.Send(ui.RecvProgressMsg{
					ReceivedBytes: -1,
					TotalBytes:    -1,
					Done:          true,
				})
			}
		}()

		if _, err := program.Run(); err != nil {
			cancel()
			return err
		}
		cancel()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(recvCmd)
}
