package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/contextdata"
	"github.com/daffahilmyf/ride-hailing/services/gateway/internal/app/responses"
)

var userHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
}

func ProxyUser(baseURL string, includeIdentity bool, internalToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if baseURL == "" {
			responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "USER_UNCONFIGURED"})
			return
		}
		target, err := buildUserURL(baseURL, c.Request.URL.Path, c.Request.URL.RawQuery)
		if err != nil {
			responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "INVALID_USER_URL"})
			return
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, target, c.Request.Body)
		if err != nil {
			responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "REQUEST_BUILD_FAILED"})
			return
		}
		copyHeaders(req.Header, c.Request.Header)
		if includeIdentity {
			userID := contextdata.GetUserID(c)
			role := contextdata.GetRole(c)
			if userID != "" {
				req.Header.Set("X-User-Id", userID)
			}
			if role != "" {
				req.Header.Set("X-Role", role)
			}
		}
		if internalToken != "" {
			req.Header.Set("X-Internal-Token", internalToken)
		}

		resp, err := userHTTPClient.Do(req)
		if err != nil {
			responses.RespondErrorCode(c, responses.CodeInternal, map[string]string{"reason": "USER_UNAVAILABLE"})
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
		c.Status(resp.StatusCode)
		_, _ = io.Copy(c.Writer, resp.Body)
	}
}

func buildUserURL(baseURL, path, rawQuery string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	parsed.Path = singleJoiningSlash(parsed.Path, path)
	parsed.RawQuery = rawQuery
	return parsed.String(), nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	default:
		return a + b
	}
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, v := range values {
			dst.Add(key, v)
		}
	}
}
