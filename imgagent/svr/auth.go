package svr

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	hutil "imgagent/httputil"
	"imgagent/pkg/logger"
)

const (
	// userInfo auth 认证后将 UserInfo 存储到 gin.Context 上下文中
	userInfoKey = "userInfo"
)

type UserInfo struct {
	SuperAdmin bool
	ID         int64
	Name       string
}

func (s *Service) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log := logger.FromGinContext(c)
		auth := c.GetHeader("Authorization")

		if auth == "" {
			hutil.AbortError(c, http.StatusUnauthorized, "authorization header required")
			return
		}
		prefix, token, ok := strings.Cut(auth, " ")
		if !ok || prefix != "Bearer" {
			hutil.AbortError(c, http.StatusUnauthorized, "invalid token")
			return
		}

		var userID int64
		log.Debugf("Get request token %s", token)
		// session token 认证方式
		userToken, err := s.db.UserToken(ctx, token)
		if err != nil {
			log.Warnf("Failed to get user token %s, err: %v", token, err)
			hutil.AbortError(c, http.StatusUnauthorized, "get token failed")
			return
		}

		log.Debug("Get user token ", userToken)
		if time.Since(userToken.ExpireDate) > 0 {
			log.Warn("Token is expired", userToken)
			hutil.AbortError(c, http.StatusForbidden, "token is expired")
			return
		}
		userID = userToken.UserID

		// 通过 userID 获取 xrobot 用户信息
		user, err := s.db.User(ctx, userID)
		if err != nil {
			log.Warnf("Failed to get user %d, err: %v", userID, err)
			hutil.AbortError(c, http.StatusUnauthorized, "get user failed")
			return
		}
		if user.Status != 1 {
			log.Warnf("User status %d not normal", user.Status)
			hutil.AbortError(c, http.StatusForbidden, "user not normal")
			return
		}

		ui := UserInfo{
			ID:   user.ID,
			Name: user.Username,
		}
		if user.SuperAdmin == 1 {
			ui.SuperAdmin = true
		}
		c.Set(userInfoKey, ui)

		c.Next()
	}
}

func (s *Service) NilAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func GetUserInfo(c *gin.Context) UserInfo {
	return c.MustGet(userInfoKey).(UserInfo)
}
