package hb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/castisdev/gcommon/clog"
)

const (
	heartBeatRequest  = int32(0)
	heartBeatResponse = int32(1)
)

// HeartBeatResponser :
type HeartBeatResponser struct {
	representativeIP string
	localIP          string
	processState     int32
}

// NewHeartBeatResponser :
func NewHeartBeatResponser(representativeIP, localIP string) *HeartBeatResponser {
	return &HeartBeatResponser{representativeIP, localIP, 0}
}

// ListenAndServe :
func (h *HeartBeatResponser) ListenAndServe(listenAddr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	readBuffer := make([]byte, 1024)
	for {
		n, addr, err := udpConn.ReadFromUDP(readBuffer)
		if err != nil {
			clog.Errorf("failed to read, %v", err)
			continue
		}
		remoteEP := ""
		if addr != nil {
			remoteEP = addr.String()
		}
		if n != 8 {
			clog.Warningf1(remoteEP, "invalid msg size, %d", n)
			continue
		}
		_, seq, err := readMsgTypeSeq(readBuffer[:n])
		if err != nil {
			clog.Errorf1(remoteEP, "%v", err)
			continue
		}

		w := new(bytes.Buffer)
		if err := binary.Write(w, binary.BigEndian, heartBeatResponse); err != nil {
			clog.Errorf1(remoteEP, "%v", err)
			continue
		}
		if err := binary.Write(w, binary.BigEndian, seq); err != nil {
			clog.Errorf1(remoteEP, "%v", err)
			continue
		}
		if err := writeString(w, h.representativeIP); err != nil {
			clog.Errorf1(remoteEP, "%v", err)
			continue
		}
		if err := binary.Write(w, binary.BigEndian, h.processState); err != nil {
			clog.Errorf1(remoteEP, "%v", err)
			continue
		}
		if err := writeString(w, h.localIP); err != nil {
			clog.Errorf1(remoteEP, "%v", err)
			continue
		}
		toWriten := w.Len()
		n, err = udpConn.WriteToUDP(w.Bytes(), addr)
		if err != nil {
			clog.Errorf1(remoteEP, "%v", err)
			continue
		}
		if n != toWriten {
			clog.Errorf1(remoteEP, "invalid write size, %d, %d", toWriten, n)
			continue
		}
	}
}

func readMsgTypeSeq(buf []byte) (int32, int32, error) {
	r := bytes.NewReader(buf)
	var msgType, seq int32
	if err := binary.Read(r, binary.BigEndian, &msgType); err != nil {
		return 0, 0, fmt.Errorf("failed to read, %v", err)
	}
	if msgType != heartBeatRequest {
		return 0, 0, fmt.Errorf("invalid msg type, %d", msgType)
	}
	if err := binary.Read(r, binary.BigEndian, &seq); err != nil {
		return 0, 0, fmt.Errorf("failed to read, %v", err)
	}
	return msgType, seq, nil
}

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.BigEndian, int32(len(s))); err != nil {
		return fmt.Errorf("failed to write string length, %v", err)
	}
	if err := binary.Write(w, binary.BigEndian, []byte(s)); err != nil {
		return fmt.Errorf("failed to write string, %v", err)
	}
	return nil
}
