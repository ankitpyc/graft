package validation

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/metadata"
	"time"
)

type TokenServiceInf interface {
	GetToken(secret string, follower string, leader string) string
	ValidateToken(secret string, tok string) (bool, error)
	ValidateClaims(claim *Claims) (bool, error)
	ValidateJWTFromContext(secret string, ctx context.Context) error
}

func GetToken(secret string, follower string, leader string) string {
	claims := &Claims{
		FollowerId: follower,
		LeaderId:   leader,
		ClusterId:  "1",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		},
	}

	awoken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := awoken.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("error :  ", err)
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

func ValidateJWTFromContext(secret string, ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing metadata")
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return fmt.Errorf("missing authorization header")
	}

	tokenStr := authHeader[0][len("Bearer "):]
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return fmt.Errorf("invalid token signature")
		}
		return fmt.Errorf("could not parse token: %v", err)
	}
	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	fmt.Printf("Authenticated request from NodeID: %s\n", claims.FollowerId)
	return nil
}
