package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pkglog "github.com/jordanp/goapp/pkg/log"
	"github.com/stretchr/testify/require"
)

func handlerPanic(_ http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	if _, ok := values["panic"]; ok {
		panic("my panic msg")
	}
}

func TestRecover(t *testing.T) {
	ts := httptest.NewServer(WithRecover(handlerPanic))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(ts.URL + "?panic=true")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestRecoverWithLogger(t *testing.T) {
	log, hook := pkglog.NewTest()
	chain := With(MakeLogger(log, pkglog.RequestNever), WithRecover)
	srv := httptest.NewServer(chain(handlerPanic))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "?panic=true")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Len(t, hook.Entries, 2)
	entries := hook.AllEntries()
	require.Equal(t, "my panic msg", entries[0].Message)
	require.Equal(t, pkglog.ErrorLevel, entries[1].Level)
	require.True(t, strings.HasPrefix(entries[1].Message, "goroutine "))
}
