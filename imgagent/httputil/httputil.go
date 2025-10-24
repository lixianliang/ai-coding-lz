package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"imgagent/pkg/middleware"
	"imgagent/proto"
)

const (
	ErrServerInternalCode = 599
)

// WriteData 成功时写入业务数据， data 为 api 返回的业务数据
func WriteData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, proto.BaseResponse{
		Code:  http.StatusOK,
		Reqid: c.MustGet(middleware.XReqID).(string),
		Data:  data,
	})
}

// AbortError 失败时返回的错误信息，code 表示业务错误码，msg 为错误信息
func AbortError(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(http.StatusOK, proto.BaseResponse{
		Code:    code,
		Message: msg,
		// xReqID 必须存在。
		Reqid: c.MustGet(middleware.XReqID).(string),
	})
}

// AbortErr 失败时返回的错误信息
func AbortErr(c *gin.Context, err error) {
	code := ErrServerInternalCode
	var msg string
	ae, ok := err.(*proto.ApiError)
	if ok {
		code = ae.Code
		msg = ae.Message
	} else {
		msg = err.Error()
	}
	c.AbortWithStatusJSON(http.StatusOK, proto.BaseResponse{
		Code:    code,
		Message: msg,
		// xReqID 必须存在。
		Reqid: c.MustGet(middleware.XReqID).(string),
	})
}

func NewApiError(code int, msg string) *proto.ApiError {
	return &proto.ApiError{
		Code:    code,
		Message: msg,
	}
}
