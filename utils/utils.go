package utils

import (
	"fmt"
	"os/exec"

	"golang.design/x/clipboard"
)

func GetClipboard() string {
	data := string(clipboard.Read(clipboard.FmtText))
	return data
}

func SetClipboard(s string) {
	clipboard.Write(clipboard.FmtText, []byte(s))
}
func PlaySound(soundPath string) {
	fmt.Println("Playing sound")
	// Run command  ffplay -v 0 -nodisp -autoexit /var/tmp/machine_pwned.mp3

	cmd := exec.Command("ffplay", "-v", "0", "-nodisp", "-autoexit", soundPath)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

}
