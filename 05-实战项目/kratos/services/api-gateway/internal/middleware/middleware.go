// Package middleware 提供网关中间件
package middleware

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"

	userV1 "services/api/user/v1"
)

// userKey 用于 context 中存储用户的 key
type userKey struct{}

// UserFromContext 从 context 中获取登录用户信息
func UserFromContext(ctx context.Context) (*userV1.User, bool) {
	u, ok := ctx.Value(userKey{}).(*userV1.User)
	return u, ok
}

// Auth 身份验证中间件
// 跳过白名单路径（如 /api/v1/auth/login）
func Auth(userClient userV1.UserServiceClient, whitelist []string, log *zap.Logger) func(http.Handler) http.Handler {
	white := make(map[string]struct{}, len(whitelist))
	for _, p := range whitelist {
		white[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := white[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"unauthorized: missing token"}`, http.StatusUnauthorized)
				return
			}
			tokStr := strings.TrimPrefix(authHeader, "Bearer ")

			u, err := userClient.ValidateToken(r.Context(), &userV1.ValidateTokenRequest{Token: tokStr})
			if err != nil {
				log.Warn("token 验证失败", zap.Error(err))
				http.Error(w, `{"error":"unauthorized: invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey{}, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Logger 简单日志中间件
func Logger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("HTTP 请求",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote", r.RemoteAddr),
			)
			next.ServeHTTP(w, r)
		})
	}
}
