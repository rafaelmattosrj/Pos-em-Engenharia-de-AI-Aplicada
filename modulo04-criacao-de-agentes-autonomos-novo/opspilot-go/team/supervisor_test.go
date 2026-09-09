package team

import (
	"strings"
	"testing"
)

func TestParsesDelegationDecision(t *testing.T) {
	decision, err := ParseDecision(`{"next":"analista","brief":"diagnostique"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision.Done {
		t.Fatal("expected not done")
	}
	if decision.Next != RoleAnalista {
		t.Fatalf("expected analista, got %s", decision.Next)
	}
	if decision.Brief != "diagnostique" {
		t.Fatalf("unexpected brief: %s", decision.Brief)
	}
}

func TestParsesDoneDecision(t *testing.T) {
	decision, err := ParseDecision(`{"next":"done","brief":"resumo final"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !decision.Done {
		t.Fatal("expected done")
	}
	if decision.Brief != "resumo final" {
		t.Fatalf("unexpected brief: %s", decision.Brief)
	}
}

func TestMalformedJSONErrors(t *testing.T) {
	_, err := ParseDecision("nao e json")
	if err == nil || !strings.Contains(err.Error(), "decisao invalida do supervisor") {
		t.Fatalf("expected 'decisao invalida do supervisor' error, got %v", err)
	}
}

func TestUnknownRoleErrors(t *testing.T) {
	_, err := ParseDecision(`{"next":"estagiario","brief":"x"}`)
	if err == nil || !strings.Contains(err.Error(), "papel desconhecido") {
		t.Fatalf("expected 'papel desconhecido' error, got %v", err)
	}
}
