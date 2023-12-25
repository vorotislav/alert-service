package middlewares

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/vorotislav/alert-service/internal/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func Hash(log *zap.Logger, key string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		ch := func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				return
			}

			reqHash := r.Header.Get("HashSHA256")
			if reqHash == "" {
				log.Debug("no hash in request header")

				h.ServeHTTP(w, r)

				return
			}

			decodeHash, err := base64.StdEncoding.DecodeString(reqHash)
			if err != nil {
				log.Info(fmt.Sprintf("cannot decode hash: %s; hash: %s", err.Error(), reqHash))

				http.Error(w, fmt.Sprintf("cannot decode hash: %s", err.Error()), http.StatusBadRequest)

				return
			}

			body, _ := io.ReadAll(r.Body)
			r.Body.Close()

			equal, err := utils.CheckHash(body, decodeHash, []byte(key))
			if err != nil {
				log.Info(fmt.Sprintf("cannot check hash: %s; hash: %s", err.Error(), decodeHash))

				http.Error(w, fmt.Sprintf("cannot check hashes: %s", err.Error()), http.StatusBadRequest)

				return
			}

			if !equal {
				log.Info("hash not equal")

				http.Error(w, "hash not equal", http.StatusBadRequest)

				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(ch)
	}
}
