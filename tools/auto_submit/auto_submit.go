package autosubmit

import (
	"fmt"
	"strings"
	"time"

	"github.com/c3l3si4n/styx/api"
	"github.com/c3l3si4n/styx/config"
	"github.com/c3l3si4n/styx/utils"
)

func isValidHex(s string) bool {
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F') {
			return false
		}
	}
	return true
}

var SubmittedFlags = make(map[string]bool)

func StartSubmitter() {
	// downlaod https://app.hackthebox.com/soundEffects/machine_pwned.mp3
	api.DownloadFile("https://app.hackthebox.com/soundEffects/machine_pwned.mp3", "/var/tmp/machine_pwned.mp3")

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {

			select {

			case <-ticker.C:
				if config.AutoFlagSubmit && config.SelectedMachineFound {
					// Get clipboard data
					clipboardData := strings.Trim(utils.GetClipboard(), " \n")
					// Check if clipboard data is a valid flag
					if len(clipboardData) == 32 {
						if isValidHex(clipboardData) {
							if _, ok := SubmittedFlags[clipboardData]; !ok {
								// Submit flag
								owned := api.SubmitFlag(clipboardData, config.SelectedMachine.Details.Id, config.SelectedMachine.Details.MachineMode)
								if owned {
									fmt.Println("Flag submitted successfully")
									// Play sound
									utils.PlaySound("/var/tmp/machine_pwned.mp3")

								}

								// Add flag to submitted flags
								SubmittedFlags[clipboardData] = true

								// Clear clipboard
							}

						}
					}
				}

			}

		}
	}()
}
