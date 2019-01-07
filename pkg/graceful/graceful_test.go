package graceful

import (
	"net"
	"syscall"
	"testing"
	"time"

	pkglog "github.com/jordanp/goapp/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestMakeListenAndServe(t *testing.T) {
	log, hook := pkglog.NewTest()
	listenAndServe := MakeListenAndServe(log, time.Second)

	time.AfterFunc(500*time.Millisecond, func() {
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	})

	if err := listenAndServe(":0", nil); err != nil {
		log.WithError(err).Error("listenAndServe")
	}

	entries := hook.AllEntries()
	require.Len(t, entries, 4)
	require.Equal(t, "starting server shutdown (signal=interrupt timeout=1s)", entries[1].Message)
}

func TestMakeListenAndServeWithPortAlreadyBound(t *testing.T) {
	log, _ := pkglog.NewTest()
	listenAndServe := MakeListenAndServe(log, time.Second)

	l, err := net.Listen("tcp", "localhost:4242")
	if err != nil {
		t.Skip("port 4242 already bound")
	}
	defer l.Close()

	err = listenAndServe(":4242", nil)
	require.Error(t, err)
	require.Equal(t, "listen", err.(*net.OpError).Op)
}
