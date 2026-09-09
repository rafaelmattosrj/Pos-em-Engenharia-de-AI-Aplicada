package framework

import (
	"testing"

	"decision-framework-tool/ahp"
	"decision-framework-tool/config"
)

func carregarConfigEPesos(t *testing.T) (*config.AmplitudeConfig, []float64) {
	t.Helper()
	cfg, err := config.Carregar()
	if err != nil {
		t.Fatal(err)
	}
	pesos := ahp.DerivarPesos(cfg.Ahp.Matriz)
	return cfg, pesos
}

func casoHipotetico(g config.Governanca) config.Caso {
	return config.Caso{
		ID: "hipotetico", Nome: "Caso hipotético", Tarefa: "tarefa",
		Scores:     map[string]float64{"p1": 0.9, "p2": 0.9, "p3": 0.9, "p4": 0.9},
		Governanca: g,
	}
}

func TestCasoSemBaseLegalEBloqueadoMesmoComScoresPerfeitos(t *testing.T) {
	_, pesos := carregarConfigEPesos(t)
	cfg, _ := config.Carregar()
	hipotetico := casoHipotetico(config.Governanca{DadoSensivelLGPD: false, BaseLegalDefinida: false, DpaAssinado: true})
	r := AvaliarCasoCompleto(hipotetico, pesos, cfg.LimiarVerde)
	if !r.BloqueadoPorGovernanca {
		t.Fatal("esperado bloqueado por governança")
	}
	if r.Aprovado {
		t.Fatal("nao deveria estar aprovado")
	}
	if r.ScoreComposto != nil {
		t.Fatalf("scoreComposto deveria ser nil quando bloqueado, obtido %v", *r.ScoreComposto)
	}
}

func TestDadoSensivelSemDpaEBloqueado(t *testing.T) {
	_, pesos := carregarConfigEPesos(t)
	cfg, _ := config.Carregar()
	hipotetico := casoHipotetico(config.Governanca{DadoSensivelLGPD: true, BaseLegalDefinida: true, DpaAssinado: false})
	r := AvaliarCasoCompleto(hipotetico, pesos, cfg.LimiarVerde)
	if !r.BloqueadoPorGovernanca {
		t.Fatal("esperado bloqueado por governança")
	}
	if len(r.MotivosGovernanca) == 0 {
		t.Fatal("esperado motivo de governança")
	}
}

func TestDadoSensivelComDpaPassaNormalmente(t *testing.T) {
	_, pesos := carregarConfigEPesos(t)
	cfg, _ := config.Carregar()
	hipotetico := casoHipotetico(config.Governanca{DadoSensivelLGPD: true, BaseLegalDefinida: true, DpaAssinado: true})
	r := AvaliarCasoCompleto(hipotetico, pesos, cfg.LimiarVerde)
	if r.BloqueadoPorGovernanca {
		t.Fatal("nao deveria estar bloqueado")
	}
	if !r.Aprovado {
		t.Fatal("esperado aprovado")
	}
}

func TestScoreExatamenteNoLimiarContaComoVerde(t *testing.T) {
	cfg, pesos := carregarConfigEPesos(t)
	r := AvaliarFramework(map[string]float64{"p1": cfg.LimiarVerde, "p2": 0.9, "p3": 0.9, "p4": 0.9}, pesos, cfg.LimiarVerde)
	if r.SinaisPorPergunta["p1"].Sinal != "VERDE" {
		t.Fatalf("esperado VERDE no limiar, obtido %s", r.SinaisPorPergunta["p1"].Sinal)
	}
	if !r.Aprovado {
		t.Fatal("esperado aprovado")
	}
}

func TestQuatroComScoreAltoAprovaSemFalha(t *testing.T) {
	cfg, pesos := carregarConfigEPesos(t)
	r := AvaliarFramework(map[string]float64{"p1": 0.9, "p2": 0.9, "p3": 0.9, "p4": 0.9}, pesos, cfg.LimiarVerde)
	if !r.Aprovado {
		t.Fatal("esperado aprovado")
	}
	if len(r.PerguntasFalhas) != 0 {
		t.Fatalf("esperado sem falhas, obtido %v", r.PerguntasFalhas)
	}
}

func TestAprovadoDeixaDecisaoDeTecnicaEmAberto(t *testing.T) {
	cfg, pesos := carregarConfigEPesos(t)
	aprovado := AvaliarFramework(map[string]float64{"p1": 0.9, "p2": 0.9, "p3": 0.9, "p4": 0.9}, pesos, cfg.LimiarVerde)
	reprovado := AvaliarFramework(map[string]float64{"p1": 0.2, "p2": 0.9, "p3": 0.9, "p4": 0.9}, pesos, cfg.LimiarVerde)
	if !aprovado.DecisaoTecnicaEmAberto {
		t.Fatal("esperado decisão técnica em aberto quando aprovado")
	}
	if reprovado.DecisaoTecnicaEmAberto {
		t.Fatal("nao esperado decisão técnica em aberto quando reprovado")
	}
}

func TestPergunta3AbaixoDoLimiarFalhaSoDado(t *testing.T) {
	cfg, pesos := carregarConfigEPesos(t)
	r := AvaliarFramework(map[string]float64{"p1": 0.9, "p2": 0.9, "p3": 0.2, "p4": 0.9}, pesos, cfg.LimiarVerde)
	if r.Aprovado {
		t.Fatal("nao esperado aprovado")
	}
	if len(r.PerguntasFalhas) != 1 || r.PerguntasFalhas[0] != 3 {
		t.Fatalf("esperado falha só em p3 (indice 3), obtido %v", r.PerguntasFalhas)
	}
	if !r.FalhaSoDado {
		t.Fatal("esperado falhaSoDado=true")
	}
}

func TestPergunta1AbaixoDoLimiarNaoEFalhaSoDado(t *testing.T) {
	cfg, pesos := carregarConfigEPesos(t)
	r := AvaliarFramework(map[string]float64{"p1": 0.2, "p2": 0.9, "p3": 0.9, "p4": 0.9}, pesos, cfg.LimiarVerde)
	if r.FalhaSoDado {
		t.Fatal("nao esperado falhaSoDado=true")
	}
}

func TestAmplitudeAutoEAprovado(t *testing.T) {
	cfg, pesos := carregarConfigEPesos(t)
	auto, err := cfg.Caso("amplitude-auto")
	if err != nil {
		t.Fatal(err)
	}
	r := AvaliarFramework(auto.Scores, pesos, cfg.LimiarVerde)
	if !r.Aprovado {
		t.Fatal("esperado amplitude-auto aprovado")
	}
}

func TestAmplitudeSaudeReprovadoSoPorDado(t *testing.T) {
	cfg, pesos := carregarConfigEPesos(t)
	saude, err := cfg.Caso("amplitude-saude-empresarial")
	if err != nil {
		t.Fatal(err)
	}
	r := AvaliarFramework(saude.Scores, pesos, cfg.LimiarVerde)
	if r.Aprovado {
		t.Fatal("nao esperado aprovado")
	}
	if !r.FalhaSoDado {
		t.Fatal("esperado falhaSoDado=true")
	}
}

func TestAmplitudeAtendimentoReprovadoEmP1EP4(t *testing.T) {
	cfg, pesos := carregarConfigEPesos(t)
	atendimento, err := cfg.Caso("amplitude-atendimento-cliente")
	if err != nil {
		t.Fatal(err)
	}
	r := AvaliarFramework(atendimento.Scores, pesos, cfg.LimiarVerde)
	if r.Aprovado {
		t.Fatal("nao esperado aprovado")
	}
	if len(r.PerguntasFalhas) != 2 || r.PerguntasFalhas[0] != 1 || r.PerguntasFalhas[1] != 4 {
		t.Fatalf("esperado falhas em [1 4], obtido %v", r.PerguntasFalhas)
	}
	if r.FalhaSoDado {
		t.Fatal("nao esperado falhaSoDado=true")
	}
	if r.SinaisPorPergunta["p3"].Sinal != "VERDE" {
		t.Fatalf("esperado p3 VERDE, obtido %s", r.SinaisPorPergunta["p3"].Sinal)
	}
	if r.Recomendacao != ContinuarPromptRAG {
		t.Fatalf("esperado recomendacao ContinuarPromptRAG, obtido %v", r.Recomendacao)
	}
}

func TestComiteNaoMudaOVereditoDosTresCasosReais(t *testing.T) {
	cfg, pesos := carregarConfigEPesos(t)
	matrizGestorProduto := [][]float64{
		{1, 1, 1.0 / 3, 0.5}, {1, 1, 1.0 / 3, 0.5}, {3, 3, 1, 2}, {2, 2, 0.5, 1},
	}
	matrizCompliance := [][]float64{
		{1, 2, 1.0 / 5, 1.0 / 3}, {0.5, 1, 1.0 / 5, 1.0 / 3}, {5, 5, 1, 3}, {3, 3, 1.0 / 3, 1},
	}
	matrizEngenharia := [][]float64{
		{1, 1, 0.5, 1}, {1, 1, 0.5, 1}, {2, 2, 1, 2}, {1, 1, 0.5, 1},
	}
	matrizAgregada := ahp.AgregarMatrizesComite([][][]float64{matrizGestorProduto, matrizCompliance, matrizEngenharia})
	pesosComite := ahp.DerivarPesos(matrizAgregada)

	for _, id := range []string{"amplitude-auto", "amplitude-saude-empresarial", "amplitude-atendimento-cliente"} {
		caso, err := cfg.Caso(id)
		if err != nil {
			t.Fatal(err)
		}
		original := AvaliarFramework(caso.Scores, pesos, cfg.LimiarVerde).Aprovado
		comite := AvaliarFramework(caso.Scores, pesosComite, cfg.LimiarVerde).Aprovado
		if comite != original {
			t.Fatalf("veredito de %s mudou entre matriz única e comitê", id)
		}
	}
}
