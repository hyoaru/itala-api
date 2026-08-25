package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	identityprovider "github.com/hyoaru/itala-api/internal/features/identity/application/port/identityprovider"
	entity "github.com/hyoaru/itala-api/internal/features/identity/domain/entity"
)

type CognitoIdentityProvider struct {
	region     string
	userPoolID string
}

func NewCognitoIdentityProvider(region string, userPoolID string) *CognitoIdentityProvider {
	return &CognitoIdentityProvider{region: region, userPoolID: userPoolID}
}

func (idp *CognitoIdentityProvider) ValidateToken(ctx context.Context, token string) (entity.Claims, error) {
	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", idp.region, idp.userPoolID)

	jwks, err := keyfunc.NewDefault([]string{issuer + "/.well-known/jwks.json"})
	if err != nil {
		return entity.Claims{}, fmt.Errorf("get cognito jwks: %w", err)
	}

	parsedToken, err := jwt.Parse(
		token,
		jwks.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(issuer),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return entity.Claims{}, identityprovider.ErrTokenExpired
		}

		return entity.Claims{}, identityprovider.ErrTokenInvalid
	}

	if !parsedToken.Valid {
		return entity.Claims{}, identityprovider.ErrTokenInvalid
	}

	sub, err := parsedToken.Claims.GetSubject()
	if err != nil || sub == "" {
		return entity.Claims{}, identityprovider.ErrTokenInvalid
	}

	return entity.Claims{Subject: sub}, nil
}
