package finance

import (
	"testing"

	"decision-framework-tool/config"
)

func TestRetornaOs4ParametrosRanqueadosPorAmplitudeDecrescente(t *testing.T) {
	cfg, err := config.Carregar()
	if err != nil {
		t.Fatal(err)
	}
	auto, err := cfg.Caso("amplitude-auto")
	if err != nil {
		t.Fatal(err)
	}
	sens := Analisar(auto.Financeiro, 0.2)
	if len(sens) != 4 {
		t.Fatalf("esperado 4 parametros, obtido %d", len(sens))
	}
	for i := 1; i < len(sens); i++ {
		if sens[i-1].Amplitude < sens[i].Amplitude {
			t.Fatalf("nao esta ranqueado por amplitude decrescente: %v", sens)
		}
	}
}
