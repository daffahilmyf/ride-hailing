package handlers

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

var notifyHTTPClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2: false,
	},
}

func StreamNotifications(notifyBaseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if notifyBaseURL == "" {
			responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "NOTIFY_UNCONFIGURED"})
			return
		}

		userID := contextdata.GetUserID(c)
		role := contextdata.GetRole(c)
		if userID == "" || role == "" {
			responses.RespondErrorCode(c, responses.CodeUnauthorized, map[string]string{"reason": "MISSING_IDENTITY"})
			return
		}

		targetURL, err := buildNotifyURL(notifyBaseURL, c.Query("ride_id"))
		if err != nil {
			responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "INVALID_NOTIFY_URL"})
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, targetURL, nil)
		if err != nil {
			responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "REQUEST_BUILD_FAILED"})
			return
		}

		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("X-User-Id", userID)
		req.Header.Set("X-Role", role)
		if traceID := contextdata.GetTraceID(c); traceID != "" {
			req.Header.Set("X-Trace-Id", traceID)
		}
		if requestID := contextdata.GetRequestID(c); requestID != "" {
			req.Header.Set("X-Request-Id", requestID)
		}
		if lastEventID := c.GetHeader("Last-Event-ID"); lastEventID != "" {
			req.Header.Set("Last-Event-ID", lastEventID)
		}

		resp, err := notifyHTTPClient.Do(req)
		if err != nil {
			responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "NOTIFY_UNAVAILABLE"})
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			if isHopByHopHeader(key) {
				continue
			}
			for _, v := range values {
				c.Writer.Header().Add(key, v)
			}
		}
		if c.Writer.Header().Get("Content-Type") == "" {
			c.Writer.Header().Set("Content-Type", "text/event-stream")
		}
		c.Status(resp.StatusCode)

		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "STREAM_UNSUPPORTED"})
			return
		}

		buf := make([]byte, 32*1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
					return
				}
				flusher.Flush()
			}
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}
		}
	}
}

func buildNotifyURL(baseURL, rideID string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	path := strings.TrimRight(parsed.Path, "/") + "/sse"
	parsed.Path = path
	if rideID != "" {
		q := parsed.Query()
		q.Set("ride_id", rideID)
		parsed.RawQuery = q.Encode()
	}
	return parsed.String(), nil
}

func isHopByHopHeader(key string) bool {
	switch http.CanonicalHeaderKey(key) {
	case "Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade":
		return true
	default:
		return false
	}
}
