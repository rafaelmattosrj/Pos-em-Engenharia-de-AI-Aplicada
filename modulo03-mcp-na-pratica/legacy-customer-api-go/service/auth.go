// Package service implementa a lógica de negócio da API legada — equivalente
// a AuthService.java e CustomerService.java.
package service

import (
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type userInfo struct {
	password string
	role     string
}

// users são os usuários em memória: username → {password, role} —
// equivalente ao Map USERS de AuthService.java.
var users = map[string]userInfo{
	"admin":  {password: "password123", role: "ADMIN"},
	"member": {password: "pass456", role: "MEMBER"},
}

// AuthService valida credenciais, gera e valida JWTs e service tokens —
// equivalente a AuthService.java.
type AuthService struct {
	JWTSecret     string
	JWTExpiration time.Duration
	mu            sync.RWMutex
	serviceTokens map[string]string // token UUID -> role
}

// NewAuthService cria um AuthService com o mapa de service tokens
// inicializado.
func NewAuthService(jwtSecret string, jwtExpiration time.Duration) *AuthService {
	return &AuthService{
		JWTSecret:     jwtSecret,
		JWTExpiration: jwtExpiration,
		serviceTokens: make(map[string]string),
	}
}

// Login valida credenciais e retorna o JWT gerado — equivalente a
// AuthService.login. O segundo retorno é false se as credenciais forem
// inválidas.
func (s *AuthService) Login(username, password string) (string, bool) {
	info, ok := users[username]
	if !ok || info.password != password {
		return "", false
	}
	return s.generateJWT(username, info.role), true
}

// CreateServiceToken gera um service token UUID e o registra com role
// ADMIN — equivalente a AuthService.createServiceToken. A validação do
// segredo (X-Super-Secret) é feita pelo handler, assim como na versão Java.
func (s *AuthService) CreateServiceToken() string {
	token := uuid.New().String()
	s.mu.Lock()
	s.serviceTokens[token] = "ADMIN"
	s.mu.Unlock()
	return token
}

// ValidateToken valida um token (service token ou JWT) e retorna a role
// associada — equivalente a AuthService.validateToken.
func (s *AuthService) ValidateToken(token string) (string, bool) {
	s.mu.RLock()
	role, ok := s.serviceTokens[token]
	s.mu.RUnlock()
	if ok {
		return role, true
	}

	claims, err := s.parseJWT(token)
	if err != nil {
		return "", false
	}
	role, _ = claims["role"].(string)
	if role == "" {
		return "", false
	}
	return role, true
}

// IsServiceToken retorna true se token for um service token registrado —
// equivalente a AuthService.isServiceToken.
func (s *AuthService) IsServiceToken(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.serviceTokens[token]
	return ok
}

func (s *AuthService) generateJWT(username, role string) string {
	claims := jwt.MapClaims{
		"sub":  username,
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(s.JWTExpiration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(s.JWTSecret))
	return signed
}

func (s *AuthService) parseJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(s.JWTSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		if err == nil {
			err = jwt.ErrTokenInvalidClaims
		}
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}
