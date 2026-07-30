package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/infrastructure/cache"
	"sipon-be/internal/modules/identity/infrastructure/principal"
)

func JWTAuth(tokenGen application.TokenGenerator, sessionStore application.SessionRevocationStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "MISSING_TOKEN",
					"message": "Authorization header is required",
				},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "Authorization header must be Bearer <token>",
				},
			})
			return
		}

		token := parts[1]
		claims, err := tokenGen.ParseAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Access token is invalid or expired",
				},
			})
			return
		}

		revoked, err := sessionStore.IsSessionRevoked(c.Request.Context(), claims.SessionID)
		if err != nil {
			slog.Warn("failed to check session revocation, allowing request (fail-open)", "error", err)
		} else if revoked {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "SESSION_REVOKED",
					"message": "Session has been revoked",
				},
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("session_id", claims.SessionID)
		c.Set("device_id", claims.DeviceID)

		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	v, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	s, _ := v.(string)
	return s
}

func GetSessionID(c *gin.Context) string {
	v, exists := c.Get("session_id")
	if !exists {
		return ""
	}
	s, _ := v.(string)
	return s
}

func GetDeviceID(c *gin.Context) string {
	v, exists := c.Get("device_id")
	if !exists {
		return ""
	}
	s, _ := v.(string)
	return s
}

func PrincipalLoader(builder *principal.Builder, principalCache *cache.RedisPrincipalCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.Next()
			return
		}

		var p *principal.Principal

		if principalCache != nil {
			cached, err := principalCache.Get(c.Request.Context(), userID)
			if err != nil {
				slog.Warn("failed to get principal from cache, building from DB (fail-open)", "error", err, "user_id", userID)
			} else if cached != nil {
				scopes := make([]principal.ScopeInfo, 0, len(cached.Scopes))
				for _, s := range cached.Scopes {
					scopes = append(scopes, principal.ScopeInfo{
						ScopeType: s.ScopeType,
						ScopeID:   s.ScopeID,
					})
				}
				p = &principal.Principal{
					UserID:      userID,
					Roles:       cached.Roles,
					Permissions: cached.Permissions,
					Scopes:      scopes,
				}
			}
		}

		if p == nil {
			var err error
			p, err = builder.Build(c.Request.Context(), userID)
			if err != nil {
				slog.Warn("failed to build principal, setting empty (fail-open)", "error", err, "user_id", userID)
				p = &principal.Principal{
					UserID:      userID,
					Roles:       []string{},
					Permissions: []string{},
					Scopes:      []principal.ScopeInfo{},
				}
			}

			if principalCache != nil {
				cacheData := &cache.PrincipalData{
					Roles:       p.Roles,
					Permissions: p.Permissions,
					Scopes:      make([]cache.Scope, 0, len(p.Scopes)),
				}
				for _, s := range p.Scopes {
					cacheData.Scopes = append(cacheData.Scopes, cache.Scope{
						ScopeType: s.ScopeType,
						ScopeID:   s.ScopeID,
					})
				}
				_ = principalCache.Set(c.Request.Context(), userID, cacheData, 0)
			}
		}

		c.Set("principal", p)
		c.Next()
	}
}

func GetPrincipal(c *gin.Context) *principal.Principal {
	v, exists := c.Get("principal")
	if !exists {
		return nil
	}
	p, _ := v.(*principal.Principal)
	return p
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "No principal loaded; authentication required",
				},
			})
			return
		}

		roleSet := make(map[string]struct{}, len(p.Roles))
		for _, r := range p.Roles {
			roleSet[r] = struct{}{}
		}

		for _, required := range roles {
			if _, ok := roleSet[required]; ok {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "INSUFFICIENT_ROLE",
				"message": fmt.Sprintf("Required one of roles: %s", strings.Join(roles, ", ")),
			},
		})
	}
}

func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "No principal loaded; authentication required",
				},
			})
			return
		}

		permSet := make(map[string]struct{}, len(p.Permissions))
		for _, perm := range p.Permissions {
			permSet[perm] = struct{}{}
		}

		for _, required := range permissions {
			if _, ok := permSet[required]; ok {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "INSUFFICIENT_PERMISSION",
				"message": fmt.Sprintf("Required one of permissions: %s", strings.Join(permissions, ", ")),
			},
		})
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole("usergod", "superadmin", "admin")
}

func UnmarshalPrincipal(raw string) (*principal.Principal, error) {
	type jsonPrincipal struct {
		UserID      string                `json:"user_id"`
		Roles       []string              `json:"roles"`
		Permissions []string              `json:"permissions"`
		Scopes      []principal.ScopeInfo `json:"scopes"`
	}

	var jp jsonPrincipal
	if err := json.Unmarshal([]byte(raw), &jp); err != nil {
		return nil, err
	}

	return &principal.Principal{
		UserID:      jp.UserID,
		Roles:       jp.Roles,
		Permissions: jp.Permissions,
		Scopes:      jp.Scopes,
	}, nil
}

func MarshalPrincipal(p *principal.Principal) (string, error) {
	type jsonPrincipal struct {
		UserID      string                `json:"user_id"`
		Roles       []string              `json:"roles"`
		Permissions []string              `json:"permissions"`
		Scopes      []principal.ScopeInfo `json:"scopes"`
	}

	jp := jsonPrincipal{
		UserID:      p.UserID,
		Roles:       p.Roles,
		Permissions: p.Permissions,
		Scopes:      p.Scopes,
	}
	data, err := json.Marshal(jp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
