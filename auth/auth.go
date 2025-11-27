package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/zackb/updog/id"
)

var (
	defaultScopes = []string{
		"domain:rw",
	}
)

const ScopesKey = "scopes"

// Service uses supplied JWK sets and expiration times to create and verify JWT tokens
type Service struct {

	// the "current" json web key
	jwKey jwk.Key

	// the set of all valid json web keys
	jwKeySet jwk.Set
	expiry   time.Duration
}

func NewAuthService(jwksPath string, expiry time.Duration) (*Service, error) {
	set, err := jwk.ReadFile(jwksPath)
	if err != nil {
		return nil, err
	}

	// find the latest key
	key, ok := set.Get(0)
	if !ok {
		return nil, errors.New("couldn't find a json web key")
	}

	return &Service{
		jwKey:    key,
		jwKeySet: set,
		expiry:   expiry,
	}, nil
}

type Token struct {
	ClientId string `json:"client_id"`
	Expiry   int64  `json:"expiry"`
}

// CreateToken creates a new JWT token for the given clientId (user ID).
// returns the signed token as a string and the expiration time as a Unix timestamp.
func (s *Service) CreateToken(clientId string) (string, int64, error) {

	// create jwt token
	token := jwt.New()

	// set attributes
	// aud
	_ = token.Set(jwt.AudienceKey, "updog")

	// sub
	_ = token.Set(jwt.SubjectKey, clientId)

	// exp
	exp := time.Now().Add(s.expiry)
	err := token.Set(jwt.ExpirationKey, exp)

	if err != nil {
		log.Println("failed setting expiration", err)
	}

	// iat
	err = token.Set(jwt.IssuedAtKey, time.Now().Unix())

	if err != nil {
		log.Println("failed setting issued", err)
	}

	jti := id.NewID()
	err = token.Set(jwt.JwtIDKey, jti)
	if err != nil {
		log.Println("failed setting jti", err)
	}
	// scopes
	err = token.Set(ScopesKey, defaultScopes)
	if err != nil {
		log.Println("failed setting scopes", err)
	}

	signed, err := jwt.Sign(token, jwa.HS256, s.jwKey)
	if err != nil {
		log.Println("failed signing token", err)
		return "", 0, err
	}
	return string(signed), exp.Unix(), nil
}

func (s *Service) ValidateToken(t string) (*Token, error) {
	log.Printf("validating token: %s", t)
	if t == "" {
		return nil, fmt.Errorf("token is empty")
	}

	token, err := jwt.Parse([]byte(t), jwt.WithKeySet(s.jwKeySet))

	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if token.Expiration().Unix() > 0 && token.Expiration().Before(now) {
		return nil, fmt.Errorf("token expired")
	}

	// check if the clientId has been revoked
	clientId := parseClientId(token)
	if clientId == "" {
		return nil, fmt.Errorf("missing client_id")
	}

	return &Token{
		ClientId: clientId,
		Expiry:   token.Expiration().Unix(),
	}, err
}

func (s *Service) IsAuthenticated(r *http.Request) *Token {

	// first check for the Authorization header (api)
	tokenStr := r.Header.Get("Authorization")

	// remove "Bearer " prefix if present
	if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
		tokenStr = tokenStr[7:]
	}

	// check for cookie otherwise (web)
	if tokenStr == "" {
		cookie, err := r.Cookie("token")
		if err != nil {
			log.Println("no token found in header or cookie", err)
			return nil
		}
		tokenStr = cookie.Value
	}

	// validate the token
	token, err := s.ValidateToken(tokenStr)
	if err != nil {
		log.Println("failed validating token", err)
		return nil
	}

	if token.ClientId == "" {
		log.Println("missing client_id in token")
		return nil
	}

	return token
}

func parseClientId(token jwt.Token) string {
	return token.Subject()
}
