package validators

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func BindAndValidate(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		return false
	}
	if err := validate.Struct(dst); err != nil {
		return false
	}
	return true
}
