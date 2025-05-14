package utils

import (
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

type DynamicWriter struct {
	mu     sync.RWMutex
	writer io.Writer
}

func Hello(name string) string {
	message := fmt.Sprintf("Hello %v", name)
	return message
}

func Logf(format string, a ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", a...)
}

func Errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", a...)
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

func ParseSignal(str string) syscall.Signal {
	switch str {
	case "TERM":
		return syscall.SIGTERM
	case "HUP":
		return syscall.SIGHUP
	case "INT":
		return syscall.SIGINT
	case "QUIT":
		return syscall.SIGQUIT
	case "KILL":
		return syscall.SIGKILL
	case "USR1":
		return syscall.SIGUSR1
	case "USR2":
		return syscall.SIGUSR2
	default:
		return syscall.SIGTERM
	}
}

func (d *DynamicWriter) Write(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.writer == nil {
		return 0, io.ErrClosedPipe
	}

	return d.writer.Write(p)
}

func (d *DynamicWriter) SetWriter(w io.Writer) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.writer = w
}
