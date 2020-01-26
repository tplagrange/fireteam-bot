package log

import (
	"fmt"
)

func Debug(m string) {
	fmt.Println("[DEBUG] " + m)
}

func Error(e error) {
	fmt.Println("[WARN] " + e.Error())
}

func Info(m string) {
	fmt.Println("[INFO] " + m)
}