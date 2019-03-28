package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"time"
)

var (
	expTime = 31556926 // 1 year
)

var (
	signingKey *rsa.PrivateKey
	verifyKey  *rsa.PublicKey
)

var (
	// ErrTokenInvalid denotes a token was not able to be validated.
	ErrTokenInvalid = errors.New("JWT Token was invalid")

	// ErrTokenExpired denotes a token's expire header (exp) has since passed.
	ErrTokenExpired = errors.New("JWT Token is expired")

	// ErrTokenMalformed denotes a token was not formatted as a JWT token.
	ErrTokenMalformed = errors.New("JWT Token is malformed")

	// ErrTokenNotActive denotes a token's not before header (nbf) is in the
	// future.
	ErrTokenNotActive = errors.New("token is not valid yet")

	// ErrUnexpectedSigningMethod denotes a token was signed with an unexpected
	// signing method.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

type UserClaims struct {
	ID       uint64 `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	jwt.StandardClaims
}

func parseToken(tokenString string) (*UserClaims, error) {
	claims := UserClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if token.Method != jwt.SigningMethodRS256 {
			return nil, ErrUnexpectedSigningMethod
		}

		return verifyKey, nil
	})
	if err != nil {
		if e, ok := err.(*jwt.ValidationError); ok {
			switch {
			case e.Errors&jwt.ValidationErrorMalformed != 0:
				// Token is malformed
				return nil, ErrTokenMalformed
			case e.Errors&jwt.ValidationErrorExpired != 0:
				// Token is expired
				return nil, ErrTokenExpired
			case e.Errors&jwt.ValidationErrorNotValidYet != 0:
				// Token is not active yet
				return nil, ErrTokenNotActive
			case e.Inner != nil:
				// report e.Inner
				return nil, e.Inner
			}
			// We have a ValidationError but have no specific Go kit error for it.
			// Fall through to return original error.
		}
		return nil, err
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	return &claims, nil
}

func userClaimFromToken(info *UserClaims) string {
	return "foobar"
}

func Init(privateKeyPath string, publicKeyPath string) error {
	signBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("could not read private key path: %v", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return fmt.Errorf("could not parse sign key: %v", err)
	}

	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("could not read public key: %v", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return fmt.Errorf("could not parse verify key: %v", err)
	}
	signingKey = privateKey
	verifyKey = publicKey

	return nil
}

func InitWithKeyPair(privateKey string, publicKey string) error {
	prvKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return fmt.Errorf("could not parse sign key: %v", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		return fmt.Errorf("could not parse verify key: %v", err)
	}
	signingKey = prvKey
	verifyKey = pubKey

	return nil
}

// VerifyToken verify JWT token that holds in context object
func VerifyToken(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return nil, err
	}
	tokenInfo, err := parseToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}
	grpc_ctxtags.Extract(ctx).Set("auth.sub", userClaimFromToken(tokenInfo))
	ctx = context.WithValue(ctx, "userID", tokenInfo.ID)
	return ctx, nil
}

// GenerateToken generates JWT token
func GenerateToken(_ context.Context, u UserClaims) (string, error) {
	claims := UserClaims{
		u.ID,
		u.Username,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(expTime)).Unix(),
			IssuedAt:  jwt.TimeFunc().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(signingKey)
}
