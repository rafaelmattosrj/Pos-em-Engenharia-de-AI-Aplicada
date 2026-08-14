package gateway

import (
	"reflect"
	"testing"
)

func TestTokenizarRemoveAcentosEPontuacao(t *testing.T) {
	got := Tokenizar("Idade mínima: 12 anos!")
	esperado := []string{"idade", "minima", "12", "anos"}
	if !reflect.DeepEqual(got, esperado) {
		t.Errorf("Tokenizar = %v, esperado %v", got, esperado)
	}
}

func TestTokenizarStringVazia(t *testing.T) {
	got := Tokenizar("")
	if len(got) != 0 {
		t.Errorf("Tokenizar(\"\") deveria devolver slice vazio, obtido %v", got)
	}
}
