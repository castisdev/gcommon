// ver >= go1.11 : SO_REUSEADDR 적용

// +build !go.old

package hutil

import (
	"net"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func dialer(localAddr net.Addr) *net.Dialer {
	control := func(network, address string, c syscall.RawConn) error {
		var err error
		c.Control(func(fd uintptr) {
			err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
			if err != nil {
				return
			}
		})
		return err
	}
	return &net.Dialer{
		LocalAddr: localAddr,
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Control:   control,
	}
}
