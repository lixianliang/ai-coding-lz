package proto

import "fmt"

// BaseResponse 响应结构体。
type BaseResponse struct {
	// Code 业务处理状态吗， 200 表示 api 业务处理成功，非 200 表示失败。
	Code int `json:"code"`
	// Message 错误信息。
	Message string `json:"message,omitempty"`
	// Reqid 请求 id。
	Reqid string `json:"reqid"`
	// Data 业务数据。
	Data any `json:"data"`
}

type ApiError struct {
	// Code 业务处理状态吗， 200 表示 api 业务处理成功，非 200 表示失败。
	Code int
	// Message 错误信息。
	Message string
}

func (e *ApiError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}
