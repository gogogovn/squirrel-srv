package version

import (
	"encoding/json"
	"log"
	"net/http"
)

// Handler handle version info request
func Handler(w http.ResponseWriter, _ *http.Request) {
	info := Version{BuildTime, Commit, Release}

	body, err := json.Marshal(info)
	if err != nil {
		log.Printf("Could not encode info data: %v", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
