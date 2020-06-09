package hutil

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/castisdev/gcommon/clog"
	"github.com/juju/ratelimit"
)

// RangeResponse :
type RangeResponse http.Response

// SupportsRange :
func (r *RangeResponse) SupportsRange() bool {
	switch r.Header.Get("Accept-Ranges") {
	case "bytes":
		return true
	case "none":
		return false
	}
	return r.StatusCode == 206
}

// GetContentLength :
func (r *RangeResponse) GetContentLength() (int64, error) {
	if !r.SupportsRange() {
		return r.ContentLength, nil
	}
	cr := r.Header.Get("Content-Range")
	toks := strings.Split(cr, "/")
	if len(toks) != 2 {
		return -1, fmt.Errorf("invalid Content-Range header: %s", cr)
	}
	if toks[1] == "*" {
		return -1, nil
	}
	len, err := strconv.ParseInt(toks[1], 10, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid Content-Range header: %s", cr)
	}
	return len, nil
}

// HTTPRange :
type HTTPRange struct {
	Start, Length int64
}

// ContentRange :
func (r HTTPRange) ContentRange(size int64) string {
	sizeStr := "*"
	if size != -1 {
		sizeStr = fmt.Sprintf("%d", size)
	}
	return fmt.Sprintf("bytes %d-%d/%s", r.Start, r.Start+r.Length-1, sizeStr)
}

// ErrNotSatisfiableRange :
var ErrNotSatisfiableRange = errors.New("Not satisfiable range")

// ParseRange parses a Range header string as per RFC 2616.
// from net/http package
func ParseRange(s string, size int64) ([]HTTPRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, fmt.Errorf("invalid range header(%s)", s)
	}
	var ranges []HTTPRange
	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = strings.TrimSpace(ra)
		if ra == "" {
			continue
		}
		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, fmt.Errorf("invalid range header(%s)", s)
		}
		start, end := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])
		var r HTTPRange
		if start == "" {
			if size == -1 {
				return nil, fmt.Errorf("invalid range header(%s)", s)
			}
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid range header(%s)", s)
			}
			if i > size {
				i = size
			}
			r.Start = size - i
			r.Length = size - r.Start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid range header(%s)", s)
			}
			if (size != -1 && i >= size) || i < 0 {
				return nil, ErrNotSatisfiableRange
			}
			r.Start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.Length = size - r.Start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.Start > i {
					return nil, fmt.Errorf("invalid range header(%s)", s)
				}
				if size != -1 && i >= size {
					i = size - 1
				}
				r.Length = i - r.Start + 1
			}
		}
		ranges = append(ranges, r)
	}
	return ranges, nil
}

// SetOriginRequestHeader :
func SetOriginRequestHeader(originReqHeader *http.Header, clientReqHeader http.Header) {
	for k, vv := range clientReqHeader {
		if k == "Host" || k == "Range" || k == "Connection" {
			continue
		}
		originReqHeader.Del(k)
		for _, v := range vv {
			originReqHeader.Add(k, v)
		}
	}
}

// RateLimitResponseWriter :
type RateLimitResponseWriter struct {
	respWriter http.ResponseWriter
	bucket     *ratelimit.Bucket
}

// NewRateLimitResponseWriter : if bucket is nil, no limit
func NewRateLimitResponseWriter(w http.ResponseWriter, bucket *ratelimit.Bucket) http.ResponseWriter {
	return &RateLimitResponseWriter{respWriter: w, bucket: bucket}
}

// Header :
func (r *RateLimitResponseWriter) Header() http.Header {
	return r.respWriter.Header()
}

// WriteHeader :
func (r *RateLimitResponseWriter) WriteHeader(code int) {
	r.respWriter.WriteHeader(code)
}

// Write :
func (r *RateLimitResponseWriter) Write(b []byte) (int, error) {
	if r.bucket != nil {
		nwrited := 0
		unit := int(r.bucket.Capacity())
		w := ratelimit.Writer(r.respWriter, r.bucket)
		for i := 0; i <= len(b)/unit; i++ {
			s := i * unit
			if s >= len(b) {
				break
			}
			e := s + unit
			if e > len(b) {
				e = len(b)
			}

			n, err := w.Write(b[s:e])
			if err != nil {
				return nwrited, err
			}
			nwrited += n
		}
		return nwrited, nil
	}
	return r.respWriter.Write(b)
}

// Flush :
func (r *RateLimitResponseWriter) Flush() {
	if f, ok := r.respWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack :
func (r *RateLimitResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := r.respWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("doesn't support hijacking")
}

// LogResponseWriter :
type LogResponseWriter struct {
	respWriter http.ResponseWriter
	reqID      string
	originKey  string
	req        *http.Request
}

// NewLogResponseWriter :
func NewLogResponseWriter(w http.ResponseWriter, reqID, originKey string, req *http.Request) http.ResponseWriter {
	return &LogResponseWriter{respWriter: w, reqID: reqID, originKey: originKey, req: req}
}

// Header :
func (w *LogResponseWriter) Header() http.Header {
	return w.respWriter.Header()
}

// WriteHeader :
func (w *LogResponseWriter) WriteHeader(code int) {
	w.respWriter.WriteHeader(code)
	if code == 200 || code == 206 {
		// no log
		return
	}

	key := fmt.Sprintf("[%s%s]", w.originKey, w.req.URL.Path)
	if clog.IsDebugEnable() {
		clog.Debugf1(w.reqID, "%s response %d %v", key, code, w.Header())
	} else {
		switch code {
		case 301, 302, 303, 307:
			clog.Infof1(w.reqID, "%s response %d, Location:%s", key, code, w.Header().Get("Location"))
		default:
			clog.Infof1(w.reqID, "%s response %d", key, code)
		}
	}
}

// Write :
func (w *LogResponseWriter) Write(b []byte) (int, error) {
	return w.respWriter.Write(b)
}

// Flush :
func (w *LogResponseWriter) Flush() {
	if f, ok := w.respWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack :
func (w *LogResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.respWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("doesn't support hijacking")
}

// HTTPClient :
type HTTPClient struct {
	*http.Client
	FollowRedirect bool
}

const redirectErrorStr = "redirect response"

// NewHTTPClient :
func NewHTTPClient(timeout time.Duration, localAddr net.Addr) *HTTPClient {
	autoRedirect := true
	return newClient(timeout, autoRedirect, localAddr)
}

// NewHTTPClientWithoutRedirect :
func NewHTTPClientWithoutRedirect(timeout time.Duration, localAddr net.Addr) *HTTPClient {
	autoRedirect := false
	return newClient(timeout, autoRedirect, localAddr)
}

// NewHTTPOverUdsClient :
func NewHTTPOverUdsClient(timeout time.Duration, sockFile string) *HTTPClient {
	autoRedirect := true
	return newClientWithUds(timeout, autoRedirect, sockFile)
}

// NewHTTPOverUdsClientWithoutRedirect :
func NewHTTPOverUdsClientWithoutRedirect(timeout time.Duration, sockFile string) *HTTPClient {
	autoRedirect := false
	return newClientWithUds(timeout, autoRedirect, sockFile)
}

func newClient(timeout time.Duration, autoRedirect bool, localAddr net.Addr) *HTTPClient {
	c := &HTTPClient{
		Client: &http.Client{
			Timeout: timeout,
			// http.DefaultTransport + (DisableKeepAlives: true)
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				Dial: (&net.Dialer{
					LocalAddr: localAddr,
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				DisableKeepAlives:     true,
			},
		},
		FollowRedirect: autoRedirect,
	}
	if c.FollowRedirect == false {
		c.CheckRedirect = c.checkRedirectError
	}
	return c
}

func newClientWithUds(timeout time.Duration, autoRedirect bool, sockFile string) *HTTPClient {
	c := &HTTPClient{
		Client: &http.Client{
			Transport: &http.Transport{
				Dial: func(__, _ string) (net.Conn, error) {
					return net.Dial("unix", sockFile)
				},
				DisableKeepAlives: true,
			},
			Timeout: timeout,
		},
		FollowRedirect: autoRedirect,
	}

	if c.FollowRedirect == false {
		c.CheckRedirect = c.checkRedirectError
	}
	return c
}

func (h *HTTPClient) isRedirect(err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), redirectErrorStr)
	}
	return false
}

func (h *HTTPClient) checkRedirectError(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		// No redirects
		return nil
	}
	return fmt.Errorf(redirectErrorStr)
}

// Do :
func (h *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	res, err := h.Client.Do(req)
	if h.isRedirect(err) {
		return res, nil
	}
	return res, err
}

// Get : wrapper of http.Get. it uses hutil.DefaultTransport()
func Get(url string) (*http.Response, error) {
	cl := NewHTTPClient(0, nil)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return cl.Do(req)
}

// Post : wrapper of http.Post. it uses hutil.DefaultTransport()
func Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	cl := NewHTTPClient(0, nil)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	return cl.Do(req)
}

// WriteJSON :
func WriteJSON(w http.ResponseWriter, r *http.Request, httpStatus int, obj interface{}) (err error) {
	var bytes []byte
	if r.FormValue("pretty") != "" {
		bytes, err = json.MarshalIndent(obj, "", "  ")
	} else {
		bytes, err = json.Marshal(obj)
	}
	if err != nil {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_, err = w.Write(bytes)
	return
}

// Query :
func Query(r *http.Request, name string) string {
	if list, ok := r.URL.Query()[name]; ok {
		return list[0]
	}
	return ""
}

// HTTPServer :
type HTTPServer struct {
	srv             *http.Server
	listener        net.Listener
	afterShutdownFn func()
}

// ServeTLS : https
func (s *HTTPServer) ServeTLS(certFile, keyFile string) error {
	if s.listener != nil {
		return s.srv.ServeTLS(s.listener, certFile, keyFile)
	}
	return s.srv.ListenAndServeTLS(certFile, keyFile)
}

// Serve :
func (s *HTTPServer) Serve() error {
	if s.listener != nil {
		return s.srv.Serve(s.listener)
	}
	return s.srv.ListenAndServe()
}

// Shutdown :
func (s *HTTPServer) Shutdown(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s.srv.Shutdown(ctx)

	if s.afterShutdownFn != nil {
		s.afterShutdownFn()
	}
}

// NewHTTPUnixSocketServer :
func NewHTTPUnixSocketServer(sockPath string, h http.Handler, shutdownFn func(),
	connStateFn func(net.Conn, http.ConnState)) (*HTTPServer, error) {
	if err := os.RemoveAll(sockPath); err != nil {
		return nil, fmt.Errorf("failed to remove unix socket file [%v], %v", sockPath, err)
	}
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen with unix domain socket [%v], %v", sockPath, err)
	}
	return &HTTPServer{
		srv:      &http.Server{Handler: h, ConnState: connStateFn},
		listener: l,
		afterShutdownFn: func() {
			os.RemoveAll(sockPath)
			if shutdownFn != nil {
				shutdownFn()
			}
		},
	}, nil
}

// NewHTTPServer :
func NewHTTPServer(addr string, h http.Handler, shutdownFn func(),
	connStateFn func(net.Conn, http.ConnState)) (*HTTPServer, error) {
	return &HTTPServer{
		srv:             &http.Server{Addr: addr, Handler: h, ConnState: connStateFn},
		listener:        nil,
		afterShutdownFn: shutdownFn,
	}, nil
}
