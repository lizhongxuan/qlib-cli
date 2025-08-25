package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic recovered: %s", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":   false,
			"code":      http.StatusInternalServerError,
			"message":   "服务器内部错误",
			"timestamp": "timestamp",
		})
	})
}