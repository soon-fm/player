package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"player/event"
	"player/run"
	"player/sockets/unix"

	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pauses the player",
	Run: func(cmd *cobra.Command, args []string) {
		defer fmt.Println("Done")
		config := unix.NewConfig()
		client := unix.NewClient()
		if err := client.Connect(config.Address()); err != nil {
			fmt.Println("Unable to connect to player:", err)
			return
		}
		defer client.Close()
		eb, err := json.Marshal(&event.Event{
			Type:    event.PauseEvent,
			Created: time.Now().UTC(),
		})
		if err != nil {
			fmt.Println("Unable to create pause event:", err)
			return
		}
		fmt.Println("Pausing Player...")
		if _, err := client.Write(eb); err != nil {
			fmt.Println("Unable to send pause event:", err)
			return
		}
		exitC := make(chan bool)
		go func() {
			defer close(exitC)
			for {
				b, err := client.Read()
				if err != nil {
					return
				}
				e := &event.Event{}
				if err := json.Unmarshal(b, e); err != nil {
					fmt.Println("error reading event:", err)
				}
				switch e.Type {
				case event.PausedEvent:
					fmt.Println("Playback paused")
					return
				case event.ErrorEvent:
					payload := &event.ErrorPayload{}
					if err := json.Unmarshal(e.Payload, payload); err != nil {
						fmt.Println("Unable to process error")
					}
					fmt.Println("Error pausing player:", payload.Error)
					return
				}
			}
		}()
		deadline := time.Second * 30
		select {
		case <-exitC:
			return
		case <-run.UntilQuit():
			return
		case <-time.After(deadline):
			fmt.Println("no response from player after", deadline)
			return
		}
	},
}
