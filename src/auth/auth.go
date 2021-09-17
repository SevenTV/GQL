package auth

import (
	"fmt"

	"github.com/SevenTV/ThreeLetterAPI/src/configure"
	"github.com/SevenTV/ThreeLetterAPI/src/utils"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

type StandardClaims = jwt.StandardClaims

func SignJWT(claim JWTClaimOptions) (string, error) {
	// Generate an unsigned token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	// Sign the token
	tokenStr, err := token.SignedString(utils.S2B(configure.Config.GetString("auth.secret")))

	return tokenStr, err
}

type JWTClaimOptions struct {
	UserID string `json:"id"`

	TokenVersion int32 `json:"ver"`
	StandardClaims
}

func VerifyJWT(token string, claim JWTClaimOptions) (*jwt.Token, error) {
	result, err := jwt.ParseWithClaims(
		token,
		claim,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Bad jwt signing method, expected HMAC but got %v", t.Header["alg"])
			}

			return utils.S2B(configure.Config.GetString("auth.secret")), nil
		},
	)

	return result, err
}

func init() {
	var err error
	googleConfig, err = getYTGConfig()
	if err != nil {
		log.WithError(err).Fatal("could not create the youtube service")
	}
}
