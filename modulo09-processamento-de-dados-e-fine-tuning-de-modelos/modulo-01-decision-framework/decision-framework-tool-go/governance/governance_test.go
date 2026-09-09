package governance

import (
	"strings"
	"testing"

	"decision-framework-tool/config"
)

func TestDadoSensivelSemDpaEBloqueado(t *testing.T) {
	r := Validar(config.Governanca{DadoSensivelLGPD: true, BaseLegalDefinida: true, DpaAssinado: false})
	if r.Aprovado {
		t.Fatal("esperado bloqueio: dado sensível sem DPA")
	}
	if len(r.Motivos) == 0 || !strings.Contains(r.Motivos[0], "DPA") {
		t.Fatalf("esperado motivo mencionando DPA, obtido %v", r.Motivos)
	}
}

func TestDadoSensivelComDpaPassaNormalmente(t *testing.T) {
	r := Validar(config.Governanca{DadoSensivelLGPD: true, BaseLegalDefinida: true, DpaAssinado: true})
	if !r.Aprovado {
		t.Fatalf("esperado aprovado, motivos=%v", r.Motivos)
	}
}

func TestCasoSemBaseLegalEBloqueado(t *testing.T) {
	r := Validar(config.Governanca{DadoSensivelLGPD: false, BaseLegalDefinida: false, DpaAssinado: true})
	if r.Aprovado {
		t.Fatal("esperado bloqueio: sem base legal definida")
	}
}

func TestOsTresCasosReaisPassamNaGovernanca(t *testing.T) {
	cfg, err := config.Carregar()
	if err != nil {
		t.Fatal(err)
	}
	for _, caso := range cfg.Casos {
		r := Validar(caso.Governanca)
		if !r.Aprovado {
			t.Fatalf("%s deveria passar na governança: %v", caso.Nome, r.Motivos)
		}
	}
}

func TestApenasSaudeEmpresarialExigeDpaDeVerdade(t *testing.T) {
	cfg, err := config.Carregar()
	if err != nil {
		t.Fatal(err)
	}
	saude, err := cfg.Caso("amplitude-saude-empresarial")
	if err != nil {
		t.Fatal(err)
	}
	if !saude.Governanca.DadoSensivelLGPD {
		t.Fatal("esperado amplitude-saude-empresarial com dado sensível")
	}
	for _, outro := range cfg.Casos {
		if outro.ID != saude.ID && outro.Governanca.DadoSensivelLGPD {
			t.Fatalf("%s não deveria ter dado sensível", outro.ID)
		}
	}
}
