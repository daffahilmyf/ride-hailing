package responses

import "github.com/gin-gonic/gin"

// WithMeta allows handlers to add extra fields to response meta.
func WithMeta(c *gin.Context, fields map[string]string) *gin.Context {
	if len(fields) == 0 {
		return c
	}
	c.Set("meta_extra", fields)
	return c
}
