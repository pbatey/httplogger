package httplogger

import (
	"net/http"
	"strconv"
)

// ResponseWriter captures the statusCode passed to WriteHeader() and length of []byte passed to Write()
type ResponseWriter interface {
	http.ResponseWriter
	StatusCode() int
	ContentLength() int
}

type responseWriter struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
}

func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	var v ResponseWriter = &responseWriter{w, 0, 0}
	return v
}

// Return statusCode passed to WriteHeader()
func (w *responseWriter) StatusCode() int {
	return w.statusCode
}

// Return Content-Length header or lenght of []bytes passed to Write()
func (w *responseWriter) ContentLength() int {
	clh := w.ResponseWriter.Header().Get("content-length")
	var cl int
	var err error
	if cl, err = strconv.Atoi(clh); err != nil {
		cl = w.contentLength
	}
	return cl
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *responseWriter) Flush() {
	z := w.ResponseWriter
	if f, ok := z.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseWriter) CloseNotify() <-chan bool {
	z := w.ResponseWriter
	return z.(http.CloseNotifier).CloseNotify()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = 200
	}
	w.contentLength += len(b)
	return w.ResponseWriter.Write(b)
}
