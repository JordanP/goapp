package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// Duration is a wrapper around time.Duration which supports correct
// marshaling to YAML and JSON. In particular, it marshals into strings, which
// can be used as map keys in json.
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	pd, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	d.Duration = pd
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

type StatusResponse struct {
	Time     time.Time `json:"time"`
	Version  string    `json:"version"`
	Hostname string    `json:"hostname"`
	Uptime   Duration  `json:"uptime"`
}

func Status(version string) func(w http.ResponseWriter, r *http.Request) {
	// If os.Hostname() returns an error, hostname will be set to "" which is not so bad so
	// no need to log or make a fuss about it.
	hostname, _ := os.Hostname()
	startTime := time.Now()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(StatusResponse{
			Time:     time.Now(),
			Version:  version,
			Hostname: hostname,
			Uptime:   Duration{time.Since(startTime)},
		})
	}
}
