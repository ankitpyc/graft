package validation

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type tokenServiceInf interface {
	GetToken(secret string, follower string, leader string) string
	ValidateToken(secret string, token string) (bool, error)
	ValidateClaims(claims *Claims) (bool, error)
}

func GetToken(secret string, follower string, leader string) string {
	claims := &Claims{
		FollowerId: follower,
		LeaderId:   leader,
		ClusterId:  "3",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		},
	}

	awoken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := awoken.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("error is ", err)
	}
	return tok
}

func ValidateToken(secret string, tok string) (bool, error) {
	claims := &Claims{}
	jwtToken, err := jwt.ParseWithClaims(tok, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return false, err
	}

	if !jwtToken.Valid {
		return false, errors.New("invalid token")
	}

	isClaimsValidated, err := ValidateClaims(claims)
	if err != nil {
		return false, err
	}
	if isClaimsValidated {
		return false, errors.New("invalid token")
	}
	return true, nil
}

func ValidateClaims(claim *Claims) (bool, error) {
	if claim.ClusterId == "" {
		return false, errors.New("invalid Claim . ClusterId is required")
	}
	return true, nil
}
