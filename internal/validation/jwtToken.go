package validation

import "github.com/dgrijalva/jwt-go"

type Claims struct {
	jwt.StandardClaims
	FollowerId string `json:"followerId"`
	LeaderId   string `json:"leaderId"`
	ClusterId  string `json:"clusterId"`
}
