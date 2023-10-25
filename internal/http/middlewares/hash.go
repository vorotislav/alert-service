package middlewares

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/vorotislav/alert-service/internal/utils"
	"io"
	"net/http"
)

func Hash(key string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		ch := func(w http.ResponseWriter, r *http.Request) {
			if key != "" {
				reqHash := r.Header.Get("HashSHA256")

				body, _ := io.ReadAll(r.Body)
				r.Body.Close()

				decodeHash, err := base64.StdEncoding.DecodeString(reqHash)
				if err != nil {
					http.Error(w, fmt.Sprintf("cannot decode hash: %s", err.Error()), http.StatusBadRequest)

					return
				}

				equal, err := utils.CheckHash(body, decodeHash, []byte(key))
				if err != nil {
					http.Error(w, fmt.Sprintf("cannot check hashes: %s", err.Error()), http.StatusBadRequest)

					return
				}

				if !equal {
					http.Error(w, "hash not equal", http.StatusBadRequest)

					return
				}

				r.Body = io.NopCloser(bytes.NewBuffer(body))

				h.ServeHTTP(w, r)
			}
		}

		return http.HandlerFunc(ch)
	}
}
