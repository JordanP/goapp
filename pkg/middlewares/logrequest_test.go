package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	pkglog "github.com/jordanp/goapp/pkg/log"
	"github.com/stretchr/testify/require"
)

func handler(w http.ResponseWriter, r *http.Request) {
	pkglog.G(r.Context()).Info("some log")
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(422)
	w.Write([]byte(`{"a": "b"}`))
}

func TestNewLogger(t *testing.T) {
	log, hook := pkglog.NewTest()
	l := MakeLogger(log, func(statusCode int) bool { return statusCode == 422 })
	srv := httptest.NewServer(l(handler))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	require.NoError(t, err)
	require.Equal(t, 422, resp.StatusCode)

	require.Len(t, hook.Entries, 2)
	entries := hook.AllEntries()
	require.Equal(t, "some log", entries[0].Message)
	require.Len(t, entries[0].Data["requestID"], 36)

	require.Equal(t, pkglog.InfoLevel, entries[1].Level)
	require.Equal(t, "422 GET "+srv.Listener.Addr().String()+"/", entries[1].Message)

	httpRequest, _ := entries[1].Data["httpRequest"].(map[string]interface{})
	require.Equal(t, int64(44), httpRequest["responseSize"]) // 44 = 32 (headers) + 2 (CRLF) + 10 (body)
	require.Equal(t, int64(57), httpRequest["requestSize"])

	require.Len(t, entries[1].Data["requestID"], 36)
}
