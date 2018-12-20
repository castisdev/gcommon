package hutil

import (
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/castisdev/gcommon/clog"
	"github.com/juju/ratelimit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangeResponse_supportsRange(t *testing.T) {
	{
		h := http.Header{}
		r := RangeResponse{StatusCode: 206, Header: h}
		assert.True(t, r.SupportsRange())
	}
	{
		h := http.Header{}
		h.Set("Accept-Ranges", "bytes")
		r := RangeResponse{StatusCode: 200, Header: h}
		assert.True(t, r.SupportsRange())
	}
	{
		h := http.Header{}
		h.Set("Accept-Ranges", "none")
		r := RangeResponse{StatusCode: 206, Header: h}
		assert.False(t, r.SupportsRange())
	}
	{
		h := http.Header{}
		r := RangeResponse{StatusCode: 200, Header: h}
		assert.False(t, r.SupportsRange())
	}
}

func TestRangeResponse_contentLength(t *testing.T) {
	{
		h := http.Header{}
		h.Set("Content-Range", "any string/3")
		r := RangeResponse{StatusCode: 206, Header: h}
		len, err := r.GetContentLength()
		assert.Nil(t, err)
		assert.Equal(t, int64(3), len)
	}

	// not support Range
	{
		h := http.Header{}
		h.Set("Accept-Ranges", "none")
		r := RangeResponse{StatusCode: 200, ContentLength: 100, Header: h}
		len, err := r.GetContentLength()
		assert.Nil(t, err)
		assert.Equal(t, int64(100), len)
	}

	// no Content-Range
	{
		h := http.Header{}
		r := RangeResponse{StatusCode: 206, Header: h}
		len, err := r.GetContentLength()
		assert.NotNil(t, err)
		assert.Equal(t, int64(-1), len)
	}

	// invalid Content-Range size
	{
		h := http.Header{}
		h.Set("Content-Range", "any string/invalid size")
		r := RangeResponse{StatusCode: 206, Header: h}
		len, err := r.GetContentLength()
		assert.NotNil(t, err)
		assert.Equal(t, int64(-1), len)
	}
}

func TestSetOriginRequestHeader(t *testing.T) {
	origin := http.Header{}
	client := http.Header{}

	origin.Set("Host", "originHost")
	origin.Set("Range", "originRange")
	origin.Set("Aaa", "originA")

	client.Set("Aaa", "a")
	client.Add("Bbb", "b1")
	client.Add("Bbb", "b2")
	client.Set("Host", "clientHost")
	client.Set("Range", "clientRange")

	SetOriginRequestHeader(&origin, client)

	assert.Equal(t, "originHost", origin.Get("Host"))
	assert.Equal(t, "originRange", origin.Get("Range"))
	assert.Equal(t, "a", origin.Get("Aaa"))
	assert.Equal(t, []string{"b1", "b2"}, origin["Bbb"])
}

func isCloseTo(x, y, tolerence time.Duration) bool {
	return math.Abs(float64(x)-float64(y)) < float64(tolerence)
}

type winfo struct {
	nwrited   int
	start     int64
	realstart int64
	end       int64
}

type fakeWriter struct {
	io.Writer
	started time.Time
	wait    time.Duration
	winfos  []winfo
}

func (f *fakeWriter) Write(b []byte) (int, error) {
	realstart := time.Since(f.started).Nanoseconds()
	if f.wait.Nanoseconds() > 0 {
		<-time.After(f.wait)
	}
	f.winfos = append(f.winfos, winfo{nwrited: len(b), realstart: realstart})
	return len(b), nil
}

func TestRateLimit_SomeWriteDelaySideEffect(t *testing.T) {
	gs := time.Now()
	bucket := ratelimit.NewBucketWithRate(10000, 10)

	quit := make(chan bool)
	go func() {
		// delay 100 ms per every write
		delay := 100 * time.Millisecond
		w := fakeWriter{winfos: make([]winfo, 0), started: gs, wait: delay}
		ratew := ratelimit.Writer(&w, bucket)
		s := time.Now()

		for i := 0; i < 10; i++ {
			buf := make([]byte, 100)

			s1 := time.Since(gs)
			ratew.Write(buf)
			last := len(w.winfos) - 1
			w.winfos[last].start = s1.Nanoseconds()
			w.winfos[last].end = time.Since(gs).Nanoseconds()
		}

		clog.Infof("A write elapsed:%d, %#v", time.Since(s).Nanoseconds(), w.winfos)
		quit <- true
	}()

	{
		// no delay
		w := fakeWriter{winfos: make([]winfo, 0), started: gs}
		ratew := ratelimit.Writer(&w, bucket)
		s := time.Now()

		for i := 0; i < 10; i++ {
			buf := make([]byte, 100)

			s1 := time.Since(gs)
			ratew.Write(buf)
			last := len(w.winfos) - 1
			w.winfos[last].start = s1.Nanoseconds()
			w.winfos[last].end = time.Since(gs).Nanoseconds()
		}
		elapsed := time.Since(s)
		clog.Infof("B write elapsed:%d, %#v", elapsed, w.winfos)

		if elapsed > 200*time.Millisecond {
			t.Errorf("elapsed expected(less than 2 sec(no side-effect)) but(%dms)", elapsed.Nanoseconds()/1000000)
		}
	}

	<-quit
	clog.Infof("total elapsed:%d", time.Since(gs).Nanoseconds())
}

func TestRateLimitResponseWriter_Write(t *testing.T) {
	// rate limit 1000 byte per sec
	bucket := ratelimit.NewBucketWithRate(1000, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// write 1000 byte
		buf := make([]byte, 1000)
		NewRateLimitResponseWriter(w, bucket).Write(buf)
	}))
	defer ts.Close()

	s := time.Now()
	_, err := http.Get(ts.URL)
	require.Nil(t, err)

	elapsed := time.Since(s)
	expected := 1 * time.Second
	tolerence := 100 * time.Millisecond
	if !isCloseTo(elapsed, expected, tolerence) {
		t.Errorf("elapsed expected(900ms < v < 1100ms) but(%d)", elapsed.Nanoseconds()/1000000)
		return
	}
}

func TestRateLimitResponseWriter_WriteBig(t *testing.T) {
	// rate limit 100000 byte per sec
	bucket := ratelimit.NewBucketWithRate(100000, 1024)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// write 100000 byte
		buf := make([]byte, 100000)
		NewRateLimitResponseWriter(w, bucket).Write(buf)
	}))
	defer ts.Close()

	s := time.Now()
	resp, err := http.Get(ts.URL)
	require.Nil(t, err)
	defer resp.Body.Close()

	_, e := ioutil.ReadAll(resp.Body)
	require.Nil(t, e)

	elapsed := time.Since(s)
	expected := 1 * time.Second
	tolerence := 100 * time.Millisecond
	if !isCloseTo(elapsed, expected, tolerence) {
		t.Errorf("elapsed expected(900ms < v < 1100ms) but(%d)", elapsed.Nanoseconds()/1000000)
		return
	}
}

func TestRateLimitResponseWriter_WriteNoLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1000)
		// bucket is nil => no limit
		NewRateLimitResponseWriter(w, nil).Write(buf)
	}))
	defer ts.Close()

	s := time.Now()
	_, err := http.Get(ts.URL)
	require.Nil(t, err)

	elapsed := time.Since(s)
	expected := 0 * time.Millisecond
	tolerence := 10 * time.Millisecond
	if !isCloseTo(elapsed, expected, tolerence) {
		t.Errorf("elapsed expected(0ms < v < 10ms) but(%d)", elapsed.Nanoseconds()/1000000)
		return
	}
}
