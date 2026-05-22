package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"SharedLink/internal/protocol"
	"SharedLink/internal/transfer"
	"SharedLink/internal/ui"
)

var recvCmd = &cobra.Command{
	Use:   "recv [ip]",
	Short: "Receive a file from LAN",
	Long: `Receive a file from a sender on the local network.

Without arguments, scans the LAN for available senders via mDNS.
With an IP argument, connects directly using the default port (53349).
With an IP:port argument, connects to a custom port.

Examples:
  sharedlink recv                    # Scan LAN and pick a sender
  sharedlink recv 192.168.1.100      # Connect directly (default port 53349)
  sharedlink recv 192.168.1.100:53349  # Connect directly with explicit port`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		var addr string

		if len(args) == 1 {
			raw := args[0]

			host, port, err := net.SplitHostPort(raw)
			if err != nil {
				host, port = raw, protocol.DefaultPortStr
			}
			addr = net.JoinHostPort(host, port)

			if strings.TrimSpace(host) == "" {
				return fmt.Errorf("无效地址 %q: 地址不能为空", raw)
			}
		} else {
			scanModel := ui.NewScanModel(ctx)
			scanProgram := tea.NewProgram(scanModel, tea.WithContext(ctx))
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

		transferCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		model := ui.NewRecvModel()
		program := tea.NewProgram(model, tea.WithContext(transferCtx))

		go func() {
			err := transfer.Receive(transferCtx, addr, func(received int64, total int64) {
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

		result, err := program.Run()
		if err != nil {
			return err
		}

		recvModel := result.(ui.RecvModel)
		if recvModel.Err != nil {
			fmt.Printf("接收失败: %v\n", recvModel.Err)
		} else if recvModel.Done {
			fmt.Printf("接收成功！文件: %s（%.1f MB）\n", recvModel.Filename, float64(recvModel.Filesize)/1024/1024)
		} else {
			fmt.Println("已取消接收")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(recvCmd)
}
