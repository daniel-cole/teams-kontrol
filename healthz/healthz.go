package healthz

import (
	"encoding/json"
	"github.com/daniel-cole/teams-kontrol/middleware"
	"net/http"
)

type healthz struct {
	health bool
}

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	healthz := &healthz{
		health: true,
	}

	health, err := json.Marshal(healthz)
	if err != nil {
		middleware.LogWithContext(ctx).Errorf("/healthz failed to respond: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	middleware.LogWithContext(ctx).Info("/healthz request")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(health)
}
