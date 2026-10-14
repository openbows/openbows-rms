package jwtutil

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID            int64    `json:"user_id"`
	StoreID           int64    `json:"store_id"`
	Username          string   `json:"username"`
	PermissionCodes   []string `json:"permission_codes"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
	expire time.Duration
}

func NewManager(secret string, expireHours int) *Manager {
	return &Manager{
		secret: []byte(secret),
		expire: time.Duration(expireHours) * time.Hour,
	}
}

func (m *Manager) Sign(userID, storeID int64, username string, permissionCodes []string) (string, error) {
	now := time.Now()
	if permissionCodes == nil {
		permissionCodes = []string{}
	}
	claims := Claims{
		UserID:          userID,
		StoreID:         storeID,
		Username:        username,
		PermissionCodes: permissionCodes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expire)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
