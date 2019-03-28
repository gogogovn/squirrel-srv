package healthz

import "net/http"

// Handler healthz is a liveness probe.
func Handler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
