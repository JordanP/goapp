package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/gbrlsnchs/jwt"
	"github.com/google/uuid"
)

type TokenManager interface {
	ParseAccessToken(signedString string) (User, error)
	GenerateAccessToken(user User) (string, error)

	ParseAdminToken(signedString string) (AdminUser, error)
	GenerateAdminToken(user AdminUser) (string, error)
}

type tokenManager struct {
	signer              jwt.Signer
	accessTokenDuration time.Duration
	adminTokenDuration  time.Duration
}

func NewTokenManager(secretKey string) (TokenManager, error) {
	if secretKey == "" {
		return nil, errors.New("empty secret key")
	}
	tokenManager := tokenManager{
		signer:              jwt.NewHS256(secretKey),
		accessTokenDuration: 5 * time.Minute,
		adminTokenDuration:  5 * time.Minute,
	}
	return &tokenManager, nil
}

type accessTokenClaims struct {
	*jwt.JWT

	Email string `json:"email"`
	Role  string `json:"role"`
}

type adminTokenClaims struct {
	*jwt.JWT
}

func newAccessTokenClaims(login, email, role string) *accessTokenClaims {
	a := accessTokenClaims{JWT: &jwt.JWT{}}
	a.Subject = login
	a.Email = email
	a.Role = role
	a.Audience = "access"
	return &a
}

func newAdminTokenClaims(login string) *adminTokenClaims {
	a := adminTokenClaims{JWT: &jwt.JWT{}}
	a.Subject = login
	a.Audience = "admin"
	return &a
}

func (t *tokenManager) GenerateAccessToken(user User) (string, error) {
	jot := newAccessTokenClaims(user.Login, user.Email, user.Role)
	t.fillGenericClaims(jot.JWT, t.accessTokenDuration)
	token, err := t.marshal(jot)
	return string(token), err
}

func (t *tokenManager) GenerateAdminToken(user AdminUser) (string, error) {
	jot := newAdminTokenClaims(user.Login)
	t.fillGenericClaims(jot.JWT, t.adminTokenDuration)
	token, err := t.marshal(jot)
	return string(token), err
}

func (t *tokenManager) ParseAccessToken(signedString string) (User, error) {
	var jot accessTokenClaims
	if err := t.unmarshal(signedString, &jot); err != nil {
		return User{}, err
	}

	return User{Login: jot.Subject, Email: jot.Email, Role: jot.Role}, nil
}

func (t *tokenManager) ParseAdminToken(signedString string) (AdminUser, error) {
	var jot adminTokenClaims
	if err := t.unmarshal(signedString, &jot); err != nil {
		return AdminUser{}, err
	}

	return AdminUser{Login: jot.Subject}, nil
}

func (t *tokenManager) fillGenericClaims(jot *jwt.JWT, expiresIn time.Duration) {
	now := time.Now()
	jot.ExpirationTime = now.Add(expiresIn).Unix()
	jot.IssuedAt = now.Unix()
	// ID == jti == can be used to implement token revocation through blacklisting
	jot.ID = uuid.New().String()
	// Issuer == iss == can be used to restrict the validity to a part of the backend or a sub-organization
	jot.Issuer = "jordanp"

	jot.SetAlgorithm(t.signer)
	// KeyID can be used to implement secret key rotation. The idea would be to unmarshal the JWT, get the `kid`
	// and fetch the associated secret from a DB. Then use that secret to verify the signature.
	jot.SetKeyID("kid")
}

func (t *tokenManager) marshal(v interface{}) ([]byte, error) {
	payload, err := jwt.Marshal(v)
	if err != nil {
		return nil, err
	}
	return t.signer.Sign(payload)
}

func (t *tokenManager) unmarshal(token string, v interface{}) error {
	payload, sig, err := jwt.Parse(token)
	if err != nil {
		return err
	}
	if err := t.signer.Verify(payload, sig); err != nil {
		return err
	}
	if err := jwt.Unmarshal(payload, v); err != nil {
		return err
	}

	switch v := v.(type) {
	case *accessTokenClaims:
		return validate(v.JWT, "access")
	case *adminTokenClaims:
		return validate(v.JWT, "admin")
	default:
		panic(fmt.Sprintf("unknown type %T", v))

	}
}

func validate(jot *jwt.JWT, audience string) error {
	now := time.Now()
	iatValidator := jwt.IssuedAtValidator(now)
	expValidator := jwt.ExpirationTimeValidator(now)
	audValidator := jwt.AudienceValidator(audience)
	issValidator := jwt.IssuerValidator("jordanp")
	if err := jot.Validate(iatValidator, expValidator, audValidator, issValidator); err != nil {
		switch err {
		case jwt.ErrIatValidation:
			return err
		case jwt.ErrExpValidation:
			return err
		case jwt.ErrAudValidation:
			return err
		case jwt.ErrIssValidation:
			return err
		default:
			return err
		}
	}
	return nil
}
