package autosubmit

import (
	"fmt"
	"strings"
	"styx/api"
	"styx/config"
	"styx/utils"
	"time"
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
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {

			select {

			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				if config.AutoFlagSubmit && config.SelectedMachineFound {
					// Get clipboard data
					clipboardData := strings.Trim(utils.GetClipboard(), " \n")
					// Check if clipboard data is a valid flag
					if len(clipboardData) == 32 {
						if isValidHex(clipboardData) {
							if _, ok := SubmittedFlags[clipboardData]; !ok {
								// Submit flag
								err := api.SubmitFlag(clipboardData, config.SelectedMachine.Details.Id, config.SelectedMachine.Details.MachineMode)
								if err != nil {
									fmt.Println("Error submitting flag: ", err)
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
