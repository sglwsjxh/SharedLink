package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"SharedLink/internal/discover"
	"SharedLink/internal/protocol"
	"SharedLink/internal/transfer"
	"SharedLink/internal/ui"
)

var sendCmd = &cobra.Command{
	Use:   "send <file>",
	Short: "Send a file over LAN",
	Long: `Start a file transfer server and wait for a receiver to connect.

The sender listens for incoming TCP connections and transfers the specified file.
The receiver can discover this sender via mDNS or connect directly using the displayed address.

Example:
  sharedlink send ./video.mp4`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("cannot access file: %w", err)
		}
		if info.IsDir() {
			return fmt.Errorf("path is a directory, not a file")
		}

		mdnsServer, err := discover.Register(protocol.DefaultPort, discover.ServiceMeta{
			FileName: info.Name(),
			FileSize: info.Size(),
		})
		if err != nil {
			return fmt.Errorf("mDNS register: %w", err)
		}
		defer mdnsServer.Shutdown()

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		model := ui.NewSendModel(info.Name(), info.Size())
		program := tea.NewProgram(model, tea.WithContext(ctx))

		go func() {
			err := transfer.Send(ctx, "", filePath, func(sent int64, total int64) {
				program.Send(ui.SentProgressMsg{
					SentBytes:  sent,
					TotalBytes: total,
				})
			})
			if err != nil {
				program.Send(ui.SentProgressMsg{Error: err})
			} else {
				program.Send(ui.SentProgressMsg{
					SentBytes:  info.Size(),
					TotalBytes: info.Size(),
					Done:       true,
				})
			}
		}()

		result, err := program.Run()
		if err != nil {
			return err
		}

		sendModel := result.(ui.SendModel)
		if sendModel.Err != nil {
			fmt.Printf("发送失败: %v\n", sendModel.Err)
		} else if sendModel.Done {
			fmt.Printf("发送成功！文件: %s（%.1f MB）\n", sendModel.Filename, float64(sendModel.Filesize)/1024/1024)
		} else {
			fmt.Println("已取消发送")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
}
