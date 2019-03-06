// +build windows

package procutil

import (
	"fmt"
	"os"
)

// RedirectStderrToFile :
func RedirectStderrToFile(file string) (*os.File, error) {
	return nil, fmt.Errorf("not supported")
}

// SetFDLimit :
func SetFDLimit(n uint64) error {
	return nil
}

// EnableCoreDump :
func EnableCoreDump() error {
	return nil
}
