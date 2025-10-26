//go:build !governance

package middleware

import "github.com/gin-gonic/gin"


func Governance() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
