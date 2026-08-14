package service

import "guardrails-seguranca/model"

// Users é a configuração de usuários com RBAC — equivalente a
// UsersConfig.java (users.json + config.ts do TypeScript original).
var Users = map[string]model.User{
	"admin": {
		Username:    "admin",
		Role:        "admin",
		Permissions: []string{"read_files", "write_files", "delete_files"},
		DisplayName: "Admin User",
	},
	"member": {
		Username:    "member",
		Role:        "member",
		Permissions: []string{},
		DisplayName: "Regular Member",
	},
}

// ResolveUser retorna o usuário cadastrado para username, ou um usuário
// "member" padrão (com o próprio username como displayName) se não
// encontrado — equivalente a users.getOrDefault(username, ...).
func ResolveUser(username string) model.User {
	if u, ok := Users[username]; ok {
		return u
	}
	return model.User{Username: username, Role: "member", Permissions: []string{}, DisplayName: username}
}
