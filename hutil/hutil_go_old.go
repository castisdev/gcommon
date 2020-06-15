// ver < go1.11 : SO_REUSEADDR 미적용
// $ go build -tags=go.old

// +build go.old

package hutil

import (
	"net"
	"time"
)

func dialer(localAddr net.Addr) *net.Dialer {
	return &net.Dialer{
		LocalAddr: localAddr,
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
}
