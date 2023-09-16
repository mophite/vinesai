package oauth2

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"time"
	"vinesai/internel/ava"
	"vinesai/internel/x"
)

type token struct {

	// 访问令牌
	AccessToken string

	// bearer
	TokenType string

	// 刷新token
	RefreshToken string

	// 过期时间
	Expiry time.Time

	// 额外数据
	Raw string

	// 时间偏移
	expiryDelta time.Duration

	// 权限范围
	scope []string
}

func (t *token) Valid() bool {
	return t != nil && t.AccessToken != "" && !t.expired()
}

// 30天
var defaultExpiryDelta = time.Hour * 24 * 30 * 12 * 100

func (t *token) expired() bool {
	if t.Expiry.IsZero() {
		return false
	}

	expiryDelta := defaultExpiryDelta
	if t.expiryDelta != 0 {
		expiryDelta = t.expiryDelta
	}
	return t.Expiry.Round(0).Add(-expiryDelta).Before(x.LocalTimeNow())
}

type tokenClaims struct {
	Timestamp string
	jwt.RegisteredClaims
}

var jwtKey = []byte("DOGQ6MNVIU9Y5J7LK0PWB1A8H2Z4ERCX")

func generateJWToken(c *ava.Context, homeId string) (string, int64) {
	expiry := jwt.NewNumericDate(x.LocalTimeNow().Add(defaultExpiryDelta))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims{
		Timestamp: x.LocalTimeNow().Format(time.RFC3339),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "vinesai",
			Subject:   "oauth2.0授权",
			Audience:  []string{homeId},
			ExpiresAt: expiry,
			NotBefore: jwt.NewNumericDate(x.LocalTimeNow()), //token在此时间之前不能被接收处理
			IssuedAt:  jwt.NewNumericDate(x.LocalTimeNow()),
			ID:        uuid.New().String(),
		},
	})

	str, err := token.SignedString(jwtKey)
	if err != nil {
		c.Errorf("generateJWToken |err=%v", err)
		return "", 0
	}

	return str, expiry.Unix()
}

var errInvalidToken = errors.New("invalid token")

func parseJWToken(token string) (*tokenClaims, error) {
	t, err := jwt.ParseWithClaims(token, &tokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := t.Claims.(*tokenClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, errInvalidToken
}
