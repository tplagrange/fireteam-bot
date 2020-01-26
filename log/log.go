package log

import (
	"fmt"
)

func Debug(m interface{}) {
	fmt.Print("[DEBUG] ")
	fmt.Print(m)
	fmt.Print("\n")
}

func Error(e error) {
	fmt.Println("[WARN] " + e.Error())
}

func Info(m string) {
	fmt.Println("[INFO] " + m)
}