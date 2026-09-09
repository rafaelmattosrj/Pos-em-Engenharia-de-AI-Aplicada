package grpo

import "testing"

func TestCandidatoIdenticoAoGabaritoRecebeRecompensaUm(t *testing.T) {
	if r := RecompensaVerificavel(Gabarito); r != 1.0 {
		t.Fatalf("esperado 1.0, obtido %v", r)
	}
}

func TestCandidatoSemJsonValidoRecebeRecompensaZero(t *testing.T) {
	if r := RecompensaVerificavel(nil); r != 0.0 {
		t.Fatalf("esperado 0.0, obtido %v", r)
	}
}

func TestUmCampoErradoEmSeisReduzRecompensaProporcionalmente(t *testing.T) {
	parcial := make(map[string]any, len(Gabarito))
	for k, v := range Gabarito {
		parcial[k] = v
	}
	parcial["status"] = "campo errado"
	esperado := 5.0 / 6.0
	if r := RecompensaVerificavel(parcial); r < esperado-1e-9 || r > esperado+1e-9 {
		t.Fatalf("esperado %v, obtido %v", esperado, r)
	}
}

func TestDiferencaDeArredondamentoNoValorAindaContaComoAcerto(t *testing.T) {
	candidato := make(map[string]any, len(Gabarito))
	for k, v := range Gabarito {
		candidato[k] = v
	}
	candidato["estimated_amount_brl"] = 8450.001
	if r := RecompensaVerificavel(candidato); r != 1.0 {
		t.Fatalf("esperado 1.0, obtido %v", r)
	}
}
