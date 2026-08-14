package service

import (
	"testing"
	"time"
)

func newTestAuthService() *AuthService {
	return NewAuthService("test-secret-at-least-32-characters-long", time.Hour)
}

func TestLogin_ValidCredentials(t *testing.T) {
	auth := newTestAuthService()

	token, ok := auth.Login("admin", "password123")
	if !ok {
		t.Fatal("esperava login bem sucedido")
	}
	if token == "" {
		t.Error("esperava token nao vazio")
	}

	role, valid := auth.ValidateToken(token)
	if !valid || role != "ADMIN" {
		t.Errorf("esperava role=ADMIN, obteve role=%q valid=%v", role, valid)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	auth := newTestAuthService()

	_, ok := auth.Login("admin", "senha errada")
	if ok {
		t.Error("esperava login recusado para senha incorreta")
	}

	_, ok = auth.Login("usuario-inexistente", "qualquer")
	if ok {
		t.Error("esperava login recusado para usuario inexistente")
	}
}

func TestServiceToken_CreateAndValidate(t *testing.T) {
	auth := newTestAuthService()

	token := auth.CreateServiceToken()
	if !auth.IsServiceToken(token) {
		t.Error("esperava que o token criado fosse reconhecido como service token")
	}

	role, valid := auth.ValidateToken(token)
	if !valid || role != "ADMIN" {
		t.Errorf("esperava role=ADMIN para service token, obteve role=%q valid=%v", role, valid)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	auth := newTestAuthService()

	_, valid := auth.ValidateToken("token-invalido-qualquer")
	if valid {
		t.Error("esperava token invalido rejeitado")
	}
}
