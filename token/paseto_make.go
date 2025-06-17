package token

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20poly1305"
)

// PasetoMaker is a PASETO token maker
type PasetoMaker struct{
	paseto *paseto.V2
	symetricKey []byte
}

// NewPasetoMaker creates a new PasetoMaker
func NewPasetoMaker(secretKey string) (Maker, error){
	if len(secretKey) < minSecretKeySize{
		return nil, fmt.Errorf("Invalid Key size: Must be exactly %d characters", chacha20poly1305.KeySize)
	}
	maker := &PasetoMaker{
		paseto: paseto.NewV2(),
		symetricKey: []byte(secretKey),
	}
	return maker, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error){
	payload, err := NewPayload(username, duration)
	if err != nil{
		return "", err
	}

	token, err := maker.paseto.Encrypt(maker.symetricKey, payload, nil)
	if err != nil{
		return "", err
	}
	return token, nil
}

// VerifyToken checks if the token is valid or not
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error){
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symetricKey, payload, nil)
	if err != nil{
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil{
		return nil, err
	}
	return payload, nil
}	

