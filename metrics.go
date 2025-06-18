package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	const templateMetrics = `
<html>
<body>
<h1>Welcome, Chirpy Admin</h1>
<p>Chirpy has been visited %d times!</p>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf([]byte{}, templateMetrics, cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
