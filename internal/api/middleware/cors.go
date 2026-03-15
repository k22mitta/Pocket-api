package middleware

import (
	"net/http"
	"strings"
)

// CORS builds the allowed-origin header from CORS_ALLOWED_ORIGIN (a single
// origin, a comma-separated list for multiple frontend deployments, or "*").
// Defaulting to "*" is fine for local dev, but production should set this to
// the real frontend origin(s) — the API is Bearer-token authenticated (no
// cookies), so a wildcard doesn't leak credentials, but it does let any site
// call the API from a logged-in user's browser using a token it stole some
// other way.
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	allowAll := allowedOrigin == "" || allowedOrigin == "*"
	var origins []string
	if !allowAll {
		for _, o := range strings.Split(allowedOrigin, ",") {
			origins = append(origins, strings.TrimSpace(o))
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin := r.Header.Get("Origin"); origin != "" {
				for _, o := range origins {
					if o == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Vary", "Origin")
						break
					}
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
