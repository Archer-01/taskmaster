package utils

import (
	"fmt"
	"os"
	"syscall"
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

func OpenLogFile(path string) (*os.File, error) {
	flag := syscall.O_WRONLY | syscall.O_APPEND | syscall.O_CREAT
	permissions := os.FileMode(0666)

	file, err := os.OpenFile(path, flag, permissions)

	if err != nil {
		return nil, err
	}
	return file, nil
}
