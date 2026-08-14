// Package model define os tipos de domínio do sistema de guardrails —
// equivalente a User.java e GuardrailResult.java.
package model

// User representa um usuário com RBAC simples — equivalente a User.java.
type User struct {
	Username    string
	Role        string
	Permissions []string
	DisplayName string
}

// GuardrailResult é o resultado da verificação de segurança —
// equivalente a GuardrailResult.java.
type GuardrailResult struct {
	Safe     bool
	Reason   string
	Analysis string
}
