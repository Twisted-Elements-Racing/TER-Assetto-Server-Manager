package servermanager

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

const terAPITokenEnvironment = "TER_API_TOKEN"

func TERAPIAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		expectedToken := strings.TrimSpace(
			os.Getenv(terAPITokenEnvironment),
		)

		if expectedToken == "" {
			http.Error(
				w,
				"TER API authentication is not configured",
				http.StatusServiceUnavailable,
			)
			return
		}

		parts := strings.Fields(
			r.Header.Get("Authorization"),
		)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {

			w.Header().Set(
				"WWW-Authenticate",
				"Bearer",
			)

			http.Error(
				w,
				http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized,
			)
			return
		}

		providedToken := parts[1]

		expectedHash := sha256.Sum256(
			[]byte(expectedToken),
		)

		providedHash := sha256.Sum256(
			[]byte(providedToken),
		)

		if subtle.ConstantTimeCompare(
			expectedHash[:],
			providedHash[:],
		) != 1 {
			w.Header().Set(
				"WWW-Authenticate",
				"Bearer",
			)

			http.Error(
				w,
				http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
