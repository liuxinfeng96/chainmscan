package handler

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const DefaultTokenSecretKey = "baas-jwt"

type MyClaims struct {
	Id   int
	Role int
	Name string
	jwt.StandardClaims
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		token := c.Request.Header.Get("token")
		if len(token) == 0 {
			FailedJSONResp("未携带token", c)
			c.Abort()
			return
		}

		claims, err := ParseToken(token)
		if err != nil {
			FailedJSONResp("会话超时，请重新登录！", c)
			c.Abort()
			return
		}

		c.Set("token", claims)
		c.Next()
	}
}

func ParseToken(token string) (*MyClaims, error) {

	t, err := jwt.ParseWithClaims(token, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(DefaultTokenSecretKey), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			return nil, errors.New(ve.Error())
		}
		return nil, errors.New("unknown error")
	}

	if claims, ok := t.Claims.(*MyClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func GenToken(id, role int, name string, expiresAt int64) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, MyClaims{
		Id:   id,
		Role: role,
		Name: name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	})

	t, err := token.SignedString([]byte(DefaultTokenSecretKey))
	if err != nil {
		return t, err
	}

	return t, nil
}
