package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	ratelimit "rate_limit/iternal/rate-limit"
	"strconv"
)

type LoginHandler struct {
	RateService ratelimit.RateLimitInterface
}

func NewLoginHandler(rateService ratelimit.RateLimitInterface) *LoginHandler {
	return &LoginHandler{
		RateService: rateService,
	}
}

func (l *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "invalid remote address", http.StatusInternalServerError)
		return
	}

	RateStatus, retryAfter, err := l.RateService.RateLimit(r.Context(), ip)

	if RateStatus == ratelimit.RateStatusAllowed {
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(map[string]string{"status": "ok"})
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, http.StatusText(status), status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf.Bytes())
		if err != nil {
			return
		}
	}
	if RateStatus == ratelimit.RateStatusError {
		status := http.StatusServiceUnavailable
		fmt.Println(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	if RateStatus == ratelimit.RateStatusRetryAfter {
		w.Header().Set("Retry-After", strconv.FormatInt(max(retryAfter, 1), 10))
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
	}
}
