package jwtwrapper

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

type JWTWrapper struct {
	secretKey []byte
	expiresIn time.Duration
}

func NewJWTWrapper(
	secretKey string,
	expiresIn time.Duration,
) *JWTWrapper {
	return &JWTWrapper{
		[]byte(secretKey),
		expiresIn,
	}
}

func (jw *JWTWrapper) GenerateJWT(customClaims map[string]string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["exp"] = time.Now().Add(jw.expiresIn)

	for k, v := range customClaims {
		claims[k] = v
	}

	tokenString, err := token.SignedString(jw.secretKey)
	if err != nil {
		return "", errors.Wrap(err, "jwtwrapper: JWTWrapper.GenerateJWT token.SignedString error")
	}

	return tokenString, nil
}

func (jw *JWTWrapper) VerifyToken(tokenString string) (map[string]string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("jwtwrapper: JWTWrapper.VerifyToken jwt.Parse error")
		}
		return jw.secretKey, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "jwtwrapper: JWTWrapper.VerifyToken jwt.Parse token expired")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		customClaims := map[string]string{}
		for k, v := range claims {
			if k == "exp" {
				continue
			}
			customClaims[k] = v.(string)
		}
		return customClaims, nil
	}

	return nil, errors.New("jwtwrapper: JWTWrapper.VerifyToken unauthorized")
}
