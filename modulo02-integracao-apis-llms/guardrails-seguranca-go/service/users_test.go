package service

import "testing"

func TestResolveUser_KnownUser(t *testing.T) {
	u := ResolveUser("admin")
	if u.Role != "admin" || u.DisplayName != "Admin User" {
		t.Errorf("usuario admin inesperado: %+v", u)
	}
}

func TestResolveUser_UnknownUserDefaultsToMember(t *testing.T) {
	u := ResolveUser("alguem-desconhecido")
	if u.Role != "member" {
		t.Errorf("esperava role=member, obteve %q", u.Role)
	}
	if u.DisplayName != "alguem-desconhecido" {
		t.Errorf("esperava displayName=username, obteve %q", u.DisplayName)
	}
}
