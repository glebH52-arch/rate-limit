package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	ratelimit "rate_limit/iternal/rate-limit"
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
	ip := r.RemoteAddr

	RateStatus, err := l.RateService.RateLimit(r.Context(), ip)

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
		status := http.StatusTooManyRequests
		http.Error(w, err.Error(), status)
		return
	}
}
