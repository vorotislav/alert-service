package middlewares

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

func CheckSenderIP(log *zap.Logger, cidr string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		ch := func(w http.ResponseWriter, r *http.Request) {
			senderIP := r.Header.Get("X-Real-IP")

			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil {
				log.Error("cannot parse cidr", zap.Error(err))

				http.Error(w, "cannot parse cidr", http.StatusInternalServerError)

				return
			}

			if !ipNet.Contains(net.ParseIP(senderIP)) {
				log.Info("sender ip is not allowed",
					zap.String("cidr", cidr),
					zap.String("sender IP", senderIP))

				http.Error(w, "sender ip is not allowed", http.StatusForbidden)

				return
			}

			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(ch)
	}
}
