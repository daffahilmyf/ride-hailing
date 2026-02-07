package validators

import "github.com/gin-gonic/gin"

func BindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		return false
	}
	return true
}
