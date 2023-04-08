package httplogger_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pbatey/httplogger"
)

var exRemoteAddr = "10.0.0.148"
var exUserAgent = "Mozilla/5.0"
var exReferrer = "https://referrer.com/path?q=query"

func TestFormatter(t *testing.T) {
	formatter := httplogger.Compile("combined") // has all tokens

	req, _ := http.NewRequest("GET", "http://bing.com/search?q=dotnet", nil)
	req.RemoteAddr = exRemoteAddr
	req.Header.Set("user-agent", exUserAgent)
	req.Header.Set("referer", exReferrer) // standard mispelling
	req.SetBasicAuth("username", "super-secret-password")
	res := httplogger.NewResponseWriter(NewMockResponseWriter())
	res.WriteHeader(200)
	res.Header().Set("content-length", "1024")
	ts, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	dur, _ := time.ParseDuration("136us")

	format, vals := formatter(res, req, ts, dur)
	expect := `%s - %s [%s] "%s %s HTTP/%s" %s %s "%s" "%s"`
	if expect != format {
		t.Fatalf(`expected "%s" but found "%s"`, expect, format)
	}
	expectVals := []string{
		exRemoteAddr,
		"username",
		"10/Oct/2000:13:55:36 +0000",
		"GET",
		"/search",
		"1.1",
		"200",
		"1024",
		exReferrer,
		exUserAgent,
	}
	if len(expectVals) != len(vals) {
		t.Fatalf(`expected %d but found %d`, len(expectVals), len(vals))
	}
	for i, val := range vals {
		if expectVals[i] != val {
			t.Fatalf(`expected "%s" at [%d] but found "%s"`, expectVals[i], i, val)
		}
	}

	// ensure fmt.Sprintf works
	received := fmt.Sprintf(format, vals[:]...)
	expect = fmt.Sprintf(`%s - username [10/Oct/2000:13:55:36 +0000] "GET /search HTTP/1.1" 200 1024 "%s" "%s"`, exRemoteAddr, exReferrer, exUserAgent)
	if expect != received {
		t.Fatalf(`expected "%s" but found "%s"`, expect, received)
	}
}

type MockResponseWriter struct {
	header map[string][]string
}

func NewMockResponseWriter() MockResponseWriter {
	header := map[string][]string{}
	return MockResponseWriter{header}
}
func (m MockResponseWriter) CloseNotify() <-chan bool    { return nil }
func (m MockResponseWriter) Flush()                      {}
func (m MockResponseWriter) Header() http.Header         { return m.header }
func (m MockResponseWriter) Write(b []byte) (int, error) { return 0, nil }
func (m MockResponseWriter) WriteHeader(statusCode int)  {}

var _ http.ResponseWriter = MockResponseWriter{}
