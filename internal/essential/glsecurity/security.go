package glsecurity

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"go.uber.org/zap"
)

const (
	ConsoleApiCookieName = "guardlight_session"
	ContextNameUserId    = "guardlight-user-id"
)

type UserRole string

const (
	Admin  UserRole = "admin"
	Member UserRole = "member"
)

type GuardlightClaims struct {
	jwt.MapClaims
	UserId string   `json:"userId"`
	Role   UserRole `json:"role"`
}

type UserTokenCredentials struct {
	UserId uuid.UUID
	Role   UserRole
}

func MakeJwtTokenForCompanion(u UserTokenCredentials) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, GuardlightClaims{
		MapClaims: jwt.MapClaims{
			"sub": u.UserId,
			"aud": u.Role,
			"exp": time.Now().Add(time.Second * time.Duration(config.Get().Console.Jwt.MaxAge)).Unix(),
			"iat": time.Now().Unix(),
		},
		UserId: u.UserId.String(),
		Role:   u.Role,
	})
	tokenString, err := accessToken.SignedString([]byte(config.Get().Console.Jwt.SigningKey))
	if err != nil {
		zap.S().Errorw("error signing token", "error", err)
		return "", err
	}

	return tokenString, nil
}

func VerifyAndGetClaimsForGuardlightToken(tkn string) (*GuardlightClaims, error) {
	token, err := jwt.ParseWithClaims(tkn, &GuardlightClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Get().Console.Jwt.SigningKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*GuardlightClaims)
	if !ok {
		return nil, err
	}
	return claims, nil
}

func GetUserIdFromContextParsed(ctx context.Context) uuid.UUID {
	id, err := uuid.Parse(ctx.Value(ContextNameUserId).(string))
	if err != nil {
		zap.S().Errorw("error parsing user id from context", "err", err)
		return uuid.Nil
	}
	return id
}

func UseGuardlightAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		token, err := ctx.Cookie(ConsoleApiCookieName)
		if err != nil || token == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		cl, err := VerifyAndGetClaimsForGuardlightToken(token)
		if err != nil {
			zap.S().Debugw("error with jwt decoding", "error", err)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		sub, err := cl.GetSubject()
		if err != nil {
			zap.S().Errorw("error with jwt subject", "error", err)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set(ContextNameUserId, sub)
		ctx.Next()
	}
}
