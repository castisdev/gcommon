package profile

import (
	"fmt"
	"net/http"
	"os"
	"runtime/pprof"
	"sync"

	"github.com/castisdev/gcommon/clog"
	"github.com/castisdev/gcommon/hutil"
)

var muCPU sync.RWMutex
var isCPUProfileProcessing bool

// StartCPUProfile :
func StartCPUProfile(filepath string) error {
	if filepath == "" {
		filepath = "cpu.prof"
	}
	muCPU.Lock()
	if isCPUProfileProcessing {
		muCPU.Unlock()
		return fmt.Errorf("cpu profile is processing (%v)", filepath)
	}
	isCPUProfileProcessing = true
	muCPU.Unlock()
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed create cpu profile file [%v] %v", filepath, err)
	}
	pprof.StartCPUProfile(f)
	return nil
}

// StopCPUProfile :
func StopCPUProfile() {
	muCPU.RLock()
	if isCPUProfileProcessing {
		pprof.StopCPUProfile()
	}
	muCPU.RUnlock()
}

// CPUProfileHandler :
func CPUProfileHandler(w http.ResponseWriter, r *http.Request) {
	cmd := hutil.Query(r, "cmd")
	filepath := hutil.Query(r, "filepath")
	clog.Debugf("CPUProfileHandler, cmd[%v] filepath[%v]", cmd, filepath)

	switch cmd {
	case "start":
		if err := StartCPUProfile(filepath); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			clog.Debugf("cpu profile started, %v", filepath)
		}
		w.WriteHeader(http.StatusCreated)
	case "stop":
		StopCPUProfile()
		w.WriteHeader(http.StatusCreated)
		clog.Debugf("cpu profile stopped, %v", filepath)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
