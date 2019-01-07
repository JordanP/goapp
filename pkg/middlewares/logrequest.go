package middlewares

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	pkglog "github.com/jordanp/goapp/pkg/log"
)

func MakeLogger(log pkglog.Logger, logRequest func(statusCode int) bool) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log := log.F("requestID", uuid.New().String())
			ctx := pkglog.WithLogger(r.Context(), log)

			httpFields := map[string]interface{}{
				"host":          r.Host,
				"requestMethod": r.Method,
				"requestUrl":    r.URL.String(),
				"proto":         r.Proto,
				"userAgent":     r.UserAgent(),
				"referer":       r.Referer(),
			}

			r2 := *r
			rcc := &readCounterCloser{r: r.Body}
			r2.Body = rcc
			w2 := &responseStats{w: w}

			start := time.Now()
			h(w2, r2.WithContext(ctx))
			httpFields["latency"] = time.Since(start).Nanoseconds()

			statusCode := w2.code
			if w2.code == 0 {
				statusCode = http.StatusOK
			}
			httpFields["status"] = statusCode
			if !logRequest(statusCode) {
				return
			}

			if rcc.err == nil && rcc.r != nil {
				// If the handler hasn't encountered an error in the Body (like EOF),
				// then consume the rest of the Body to provide an accurate rcc.n.
				io.Copy(ioutil.Discard, rcc)
			}

			requestBodySize := rcc.n
			httpFields["requestSize"] = headerSize(r.Header) + requestBodySize

			responseHeaderSize, responseBodySize := w2.size()
			httpFields["responseSize"] = responseHeaderSize + responseBodySize

			log = log.F("httpRequest", httpFields)
			log.Infof("%d %s %s%s", httpFields["status"], r.Method, r.Host, r.URL.Path)
		}
	}
}

type readCounterCloser struct {
	r   io.ReadCloser
	n   int64
	err error
}

func (rcc *readCounterCloser) Read(p []byte) (n int, err error) {
	if rcc.err != nil {
		return 0, rcc.err
	}
	n, rcc.err = rcc.r.Read(p)
	rcc.n += int64(n)
	return n, rcc.err
}

func (rcc *readCounterCloser) Close() error {
	rcc.err = errors.New("read from closed reader")
	return rcc.r.Close()
}

type writeCounter int64

func (wc *writeCounter) Write(p []byte) (n int, err error) {
	*wc += writeCounter(len(p))
	return len(p), nil
}

func headerSize(h http.Header) int64 {
	var wc writeCounter
	h.Write(&wc)
	return int64(wc) + 2 // for CRLF
}

type responseStats struct {
	w     http.ResponseWriter
	hsize int64
	wc    writeCounter
	code  int
}

func (r *responseStats) Header() http.Header {
	return r.w.Header()
}

func (r *responseStats) WriteHeader(statusCode int) {
	if r.code != 0 {
		return
	}
	r.hsize = headerSize(r.w.Header())
	r.w.WriteHeader(statusCode)
	r.code = statusCode
}

func (r *responseStats) Write(p []byte) (n int, err error) {
	if r.code == 0 {
		r.WriteHeader(http.StatusOK)
	}
	n, err = r.w.Write(p)
	r.wc.Write(p[:n])
	return
}

func (r *responseStats) size() (hdr, body int64) {
	if r.code == 0 {
		return headerSize(r.w.Header()), 0
	}
	// Use the header size from the time WriteHeader was called.
	// The Header map can be mutated after the call to add HTTP Trailers,
	// which we don't want to count.
	return r.hsize, int64(r.wc)
}
