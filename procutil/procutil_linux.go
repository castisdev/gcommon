// +build !windows

package procutil

import (
	"fmt"
	"math"
	"os"
	"syscall"
	"time"
)

// RedirectStderrToFile :
func RedirectStderrToFile(file string) (*os.File, error) {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	str := fmt.Sprintf("\nCheckPoint-%v =====================================================================\n\n\n",
		time.Now().Format("2006-01-02,15:04:05.000000"))

	if _, err := f.WriteString(str); err != nil {
		f.Close()
		return nil, err
	}
	if err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd())); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

// SetFDLimit :
func SetFDLimit(n uint64) error {
	var rlimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit)
	if err != nil {
		return err
	}
	rlimit.Max = n
	rlimit.Cur = n
	return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit)
}

// EnableCoreDump :
func EnableCoreDump() error {
	var rlimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_CORE, &rlimit)
	if err != nil {
		return err
	}
	rlimit.Max = math.MaxUint64
	rlimit.Cur = rlimit.Max
	return syscall.Setrlimit(syscall.RLIMIT_CORE, &rlimit)
}
