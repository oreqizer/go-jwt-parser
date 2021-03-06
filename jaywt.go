package jaywt

import (
	"errors"
	"fmt"
	"gopkg.in/dgrijalva/jwt-go.v3"
	"net/http"
	"strings"
)

// TokenExtractor is a function retrieving the raw token string from a request.
type TokenExtractor func(r *http.Request) (string, error)

// Options determine the behavior of the checking functions.
type Options struct {
	// Function that will return the Key to the JWT, public key or shared secret.
	// Defaults to nil.
	Keyfunc jwt.Keyfunc
	// Function that will extract the JWT from the request.
	// Defaults to 'Authorization' header being of the form 'Bearer <token>'
	Extractor TokenExtractor
	// Which algorithm to use.
	// Defaults to jwt.SigningMethodHS256
	SigningMethod jwt.SigningMethod
}

// Core is the main structure which provides an interface for checking the token.
type Core struct {
	Options *Options
}

// New returns a new Core with the given options.
// It supplies default options for some fields (check Options type for details).
func New(o *Options) *Core {
	if o.Extractor == nil {
		o.Extractor = FromAuthHeader
	}

	if o.SigningMethod == nil {
		o.SigningMethod = jwt.SigningMethodHS256
	}

	return &Core{o}
}

// FromAuthHeader is the default extractor. It expects the 'Authorization' header
// to be in the form 'Bearer <token>'. If the header is non-existent or empty,
// it returns an empty string. Otherwise, if successful, returns the token part.
func FromAuthHeader(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", nil // No error, just no token
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("Authorization header format must be 'Bearer <token>'")
	}

	return parts[1], nil
}

// Get extracts and validates the JWT token from the request. It returns
// the parsed token, if successful.
func (m *Core) Get(r *http.Request) (*jwt.Token, error) {
	// Extract token
	raw, err := m.rawToken(r)
	if err != nil {
		return nil, err
	}

	// Parse token
	token, err := jwt.Parse(raw, m.Options.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("Error parsing token: %v", err)
	}

	// Check if token is valid
	if err = m.validateToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

// GetWithClaims extracts and validates the JWT token from the request,
// as well as the supplied claims. It returns the parsed token with the
// supplied claims, if successful.
func (m *Core) GetWithClaims(r *http.Request, claims jwt.Claims) (*jwt.Token, error) {
	// Extract token
	raw, err := m.rawToken(r)
	if err != nil {
		return nil, err
	}

	// Parse token
	token, err := jwt.ParseWithClaims(raw, claims, m.Options.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("Error parsing token: %v", err)
	}

	// Get if token is valid
	if err = m.validateToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

// Helper functions
// ---

func (m *Core) rawToken(r *http.Request) (string, error) {
	// Extract token
	raw, err := m.Options.Extractor(r)
	if err != nil {
		return "", fmt.Errorf("Error extracting token: %v", err)
	}

	// Check if token is present
	if raw == "" {
		return "", errors.New("Token not found")
	}

	return raw, nil
}

func (m *Core) validateToken(token *jwt.Token) error {
	// Verify hashing algorithm
	if alg := m.Options.SigningMethod.Alg(); alg != token.Header["alg"] {
		return fmt.Errorf("Invalid token algorithm. Wanted %s, got %s", alg, token.Header["alg"])
	}

	return nil
}
