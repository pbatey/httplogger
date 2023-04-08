package httplogger

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var tokenRegexp = regexp.MustCompile(`:([-\w]{2,})(?:\[([^\]]+)\])?`) // :<token> or :<token>[<arg>]
var defaultTimeFormat = "web"
var timeFormats = map[string]string{
	"clf": "10/Oct/2000:13:55:36 +0000",    // common log format
	"iso": "2000-10-10T13:55:36.000Z",      // ISO 8601
	"web": "Tue, 10 Oct 2000 13:55:36 GMT", // RFC 1123
}

var defaultLogFormat = "dev"
var logFormats = map[string]string{
	"combined": `:remote-addr - :remote-user [:date[clf]] ":method :url HTTP/:http-version" :status :res[content-length] ":referrer" ":user-agent"`,
	"common":   `:remote-addr - :remote-user [:date[clf]] ":method :url HTTP/:http-version" :status :res[content-length]`,
	"dev":      "\x1b[0m:method :url :status[clr] :response-time ms - :res[content-length]",
	"short":    `:remote-addr :remote-user :method :url HTTP/:http-version :status :res[content-length] - :response-time ms`,
	"tiny":     `:method :url :status :res[content-length] - :response-time ms`,
	"default":  `:remote-addr - :remote-user [:date] ":method :url HTTP/:http-version" :status :res[content-length] ":referrer" ":user-agent"`,
}

type FormatFunc func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration) (string, []any)
type TokenFunc func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string

func colorStatus(status int, s string) string {
	if status >= 500 {
		return fmt.Sprintf("\x1b[31m%s\x1b[0m", s) // server error - red
	}
	if status >= 400 {
		return fmt.Sprintf("\x1b[33m%s\x1b[0m", s) // client error - yellow
	}
	if status >= 300 {
		return fmt.Sprintf("\x1b[36m%s\x1b[0m", s) // redirect - cyan
	}
	if status >= 200 {
		return fmt.Sprintf("\x1b[32m%s\x1b[0m", s) // success - green
	}
	return s
}

func Compile(template string) FormatFunc {
	if found, ok := logFormats[template]; ok {
		template = found
	}
	matches := tokenRegexp.FindAllStringSubmatch(template, -1)
	format := tokenRegexp.ReplaceAllString(template, "%s")
	var funcs []TokenFunc
	var args []string
	for _, v := range matches {
		funcs = append(funcs, tokenFuncs[v[1]])
		args = append(args, v[2])
	}
	return func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration) (string, []any) {
		var vals []any
		for i, f := range funcs {
			val := f(res, req, ts, dur, args[i])
			if val == "" {
				val = "-"
			}
			vals = append(vals, val)
		}
		return format, vals
	}
}

var tokenFuncs = map[string]TokenFunc{
	"remote-addr": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		return req.RemoteAddr
	},
	"remote-user": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		user, _, ok := req.BasicAuth()
		if ok {
			return user
		}
		return ""
	},
	"date": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		f, ok := timeFormats[arg]
		if !ok {
			f = timeFormats[defaultTimeFormat]
		}
		return ts.Format(f)
	},
	"method": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		return req.Method
	},
	"url": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		val := ""
		if req.URL != nil {
			val = req.URL.Path
		}
		return val
	},
	"http-version": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		val := req.Proto
		if len(req.Proto) >= 5 && req.Proto[:5] == "HTTP/" {
			val = req.Proto[5:]
		}
		return val
	},
	"status": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		statusCode := res.StatusCode()
		val := "-"
		if statusCode != 0 {
			val = strconv.Itoa(statusCode)
		}
		if arg == "clr" {
			val = colorStatus(statusCode, val)
		}
		return val
	},
	"req": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		var val = ""
		if vals, ok := req.Header[http.CanonicalHeaderKey(arg)]; ok {
			val = strings.Join(vals, ", ")
		}
		return val
	},
	"res": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		val := ""
		if vals, ok := res.Header()[http.CanonicalHeaderKey(arg)]; ok {
			val = strings.Join(vals, ", ")
		}
		return val
	},
	"referrer": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		return req.Referer()
	},
	"user-agent": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		return req.UserAgent()
	},
	"response-time": func(res ResponseWriter, req *http.Request, ts time.Time, dur time.Duration, arg string) string {
		var digits int
		var err error
		if digits, err = strconv.Atoi(arg); err != nil {
			digits = 3
		}
		if digits > 0 {
			f := float64((dur.Milliseconds()*1000)+dur.Microseconds()) / 1000
			return strconv.FormatFloat(f, 'f', digits, 64)
		}
		return strconv.FormatInt(dur.Milliseconds(), 10)
	},
}

func AddToken(token string, f TokenFunc) {
	tokenFuncs[token] = f
}

func New(next http.Handler, format string) http.Handler {
	formatter := Compile(format)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		now := time.Now()
		res := NewResponseWriter(w)
		next.ServeHTTP(res, req)
		format, args := formatter(res, req, now, time.Since(now))
		log.Printf(format, args[:]...)
	})
}
