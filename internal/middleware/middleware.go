package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"openbows-rms/internal/pkg/jwtutil"
	"openbows-rms/internal/response"
)

const (
	CtxUserID          = "user_id"
	CtxStoreID         = "store_id"
	CtxUsername        = "username"
	CtxPermissionCodes = "permission_codes"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		response.Fail(c, 500, "服务器内部错误")
		c.Abort()
	})
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Writer.Header().Set("X-Request-ID", id)
		c.Set("request_id", id)
		c.Next()
	}
}

func JWTAuth(jm *jwtutil.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			response.Fail(c, 401, "未登录或令牌无效")
			c.Abort()
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		claims, err := jm.Parse(tokenStr)
		if err != nil {
			response.Fail(c, 401, "未登录或令牌无效")
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxStoreID, claims.StoreID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxPermissionCodes, claims.PermissionCodes)
		c.Next()
	}
}

// RequirePermission 要求具备任一权限码
func RequirePermission(codes ...string) gin.HandlerFunc {
	need := map[string]struct{}{}
	for _, c := range codes {
		need[c] = struct{}{}
	}
	return func(c *gin.Context) {
		if len(need) == 0 {
			c.Next()
			return
		}
		raw, _ := c.Get(CtxPermissionCodes)
		have, _ := raw.([]string)
		for _, p := range have {
			if _, ok := need[p]; ok {
				c.Next()
				return
			}
		}
		response.Fail(c, 403, "无权限")
		c.Abort()
	}
}

func GetUserID(c *gin.Context) int64 {
	v, _ := c.Get(CtxUserID)
	id, _ := v.(int64)
	return id
}

func GetStoreID(c *gin.Context) int64 {
	v, _ := c.Get(CtxStoreID)
	id, _ := v.(int64)
	return id
}

func GetUsername(c *gin.Context) string {
	v, _ := c.Get(CtxUsername)
	s, _ := v.(string)
	return s
}
