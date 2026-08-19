package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/azhai/gopaper/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthGuard struct {
	secret    []byte
	tokenTTL  time.Duration
	failCount map[string]int
	lockUntil map[string]time.Time
	mu        sync.Mutex
	admin     model.AdminConfig
}

func NewAuthGuard(secret string, ttl time.Duration, admin model.AdminConfig) *AuthGuard {
	return &AuthGuard{
		secret:    []byte(secret),
		tokenTTL:  ttl,
		failCount: make(map[string]int),
		lockUntil: make(map[string]time.Time),
		admin:     admin,
	}
}

func (ag *AuthGuard) Login(username, password string) (*model.LoginResponse, error) {
	if ag.IsLocked(username) {
		return nil, fmt.Errorf("账户已锁定，请15分钟后重试")
	}

	if username != ag.admin.USERNAME {
		ag.recordFailure(username)
		return nil, fmt.Errorf("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(ag.admin.PASSWORD), []byte(password)); err != nil {
		ag.recordFailure(username)
		return nil, fmt.Errorf("用户名或密码错误")
	}

	ag.resetFailure(username)

	now := time.Now()
	expireAt := now.Add(ag.tokenTTL)

	claims := jwt.MapClaims{
		"username": username,
		"iat":      now.Unix(),
		"exp":      expireAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(ag.secret)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	return &model.LoginResponse{
		Token:    tokenString,
		ExpireAt: expireAt,
	}, nil
}

func (ag *AuthGuard) Validate(tokenString string) (*model.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return ag.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	username, _ := claims["username"].(string)
	iat, _ := claims["iat"].(float64)
	exp, _ := claims["exp"].(float64)

	return &model.TokenClaims{
		Username: username,
		IssuedAt: int64(iat),
		ExpireAt: int64(exp),
	}, nil
}

func (ag *AuthGuard) IsLocked(username string) bool {
	ag.mu.Lock()
	defer ag.mu.Unlock()

	until, ok := ag.lockUntil[username]
	if !ok {
		return false
	}
	if time.Now().After(until) {
		delete(ag.lockUntil, username)
		delete(ag.failCount, username)
		return false
	}
	return true
}

func (ag *AuthGuard) recordFailure(username string) {
	ag.mu.Lock()
	defer ag.mu.Unlock()

	ag.failCount[username]++
	if ag.failCount[username] >= 5 {
		ag.lockUntil[username] = time.Now().Add(15 * time.Minute)
	}
}

func (ag *AuthGuard) resetFailure(username string) {
	ag.mu.Lock()
	defer ag.mu.Unlock()

	delete(ag.failCount, username)
	delete(ag.lockUntil, username)
}

func (ag *AuthGuard) EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(401, model.ErrorResponse{
					Code:    40101,
					Message: "未提供认证凭证",
				})
			}

			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				return c.JSON(401, model.ErrorResponse{
					Code:    40101,
					Message: "认证格式错误",
				})
			}

			tokenString := authHeader[7:]
			claims, err := ag.Validate(tokenString)
			if err != nil {
				return c.JSON(401, model.ErrorResponse{
					Code:    40101,
					Message: "凭证无效或已过期",
				})
			}

			c.Set("username", claims.Username)
			return next(c)
		}
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
