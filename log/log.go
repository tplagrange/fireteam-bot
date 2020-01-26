package log

import (
	"fmt"
)

func Info(m string) {
	fmt.Println("[INFO] " + m)
}

func Error(e error) {
	fmt.Println("[WARN] " + e.Error())
}