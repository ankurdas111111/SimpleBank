package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretKeySize =  32

//JWT maker is a JSON web based token maker 
type JWTMaker struct{
     secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error){
	if(len(secretKey) < minSecretKeySize){
		return nil, fmt.Errorf("Invalid Key size: Must be atleast %d 32 characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error){
	payload,err := NewPayload(username, duration)
	if err != nil{
		return "",err
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}
// VerifyToken checks if the token is valid or not
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error){
	keyFunc := func(token *jwt.Token) (interface{}, error ){
		_,ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil{
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken){
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok{
		return nil, ErrInvalidToken
	}
	
	// CRITICAL FIX: Add payload validation like PASETO has
	// This was missing and causing silent failures for expired tokens!
	err = payload.Valid()
	if err != nil{
		return nil, err
	}
	
	return payload, nil
}