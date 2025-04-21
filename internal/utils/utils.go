package utils

import (
	"fmt"
	"os"
)

func Hello(name string) string {
	message := fmt.Sprintf("Hello %v", name)
	return message
}

func Logf(format string, a ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", a...)
}

func Errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", a...)
}
