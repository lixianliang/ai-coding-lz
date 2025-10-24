package middleware

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"imgagent/pkg/logger"
)

const (
	XReqID = "X-Reqid"
)

func NewRouter(writer io.Writer) *gin.Engine {
	router := gin.New()
	// 自定义访问日志格式。
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s %s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
				param.Keys[XReqID].(string),
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		},
		Output: writer,
	}))
	router.Use(Logger())
	router.Use(Cors())
	router.Use(gin.Recovery())
	return router
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头中获取 XReqID，没有则重新生成
		reqID := c.Request.Header.Get(XReqID)
		if reqID == "" {
			uid := uuid.New()
			reqID = hex.EncodeToString(uid[:])
		}

		// 将 XReqID、log 设置到上下文中
		c.Set(XReqID, reqID)
		c.Writer.Header().Set(XReqID, reqID)
		// 将 reqid 设置到 log 中
		log := logger.NewLogger(reqID)
		c.Set(logger.ReqLogger, log)
		ctx := context.WithValue(c.Request.Context(), logger.LoggerKey, log)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "HEAD, OPTIONS, PUT, GET, POST, PATCH, DELETE")
		allowHeader := c.Request.Header.Get("Access-Control-Request-Headers")
		if allowHeader != "" {
			c.Writer.Header().Set("Access-Control-Allow-Headers", c.Request.Header.Get("Access-Control-Request-Headers"))
		}
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// 如果是预检请求（OPTIONS方法），直接返回204
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
