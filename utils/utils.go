package utils

import "golang.design/x/clipboard"

func GetClipboard() string {
	data := string(clipboard.Read(clipboard.FmtText))
	return data
}

func SetClipboard(s string) {
	clipboard.Write(clipboard.FmtText, []byte(s))
}
