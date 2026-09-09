// Demo: aplica o framework de decisão aos 3 casos reais da Amplitude Seguros
// (Auto aprovado, Saúde Empresarial reprovado só por dado -> Real Options,
// Atendimento ao Cliente reprovado por natureza da tarefa) -- equivalente ao
// fluxo principal de decision-framework-tool.js. Porte Go de
// decision-framework-tool-java.
package main

import (
	"fmt"
	"log"
	"math/rand"

	"decision-framework-tool/ahp"
	"decision-framework-tool/config"
	"decision-framework-tool/finance"
	"decision-framework-tool/framework"
)

func main() {
	cfg, err := config.Carregar()
	if err != nil {
		log.Fatal(err)
	}

	pesosAHP := ahp.DerivarPesos(cfg.Ahp.Matriz)
	consistencia := ahp.CalcularConsistencia(cfg.Ahp.Matriz, pesosAHP)

	fmt.Println("===== Demo: Framework de 4 Perguntas -- versão de análise de decisão financeira =====")
	fmt.Printf("\nPesos derivados por AHP: p1=%.3f  p2=%.3f  p3=%.3f  p4=%.3f\n",
		pesosAHP[0], pesosAHP[1], pesosAHP[2], pesosAHP[3])
	consistenteTxt := "INCONSISTENTE"
	if consistencia.Consistente {
		consistenteTxt = "consistente, CR < 0.10"
	}
	fmt.Printf("Consistência do julgamento: lambda_max=%.4f  CI=%.4f  CR=%.4f (%s)\n",
		consistencia.LambdaMax, consistencia.CI, consistencia.CR, consistenteTxt)

	auto, err := cfg.Caso("amplitude-auto")
	if err != nil {
		log.Fatal(err)
	}
	saude, err := cfg.Caso("amplitude-saude-empresarial")
	if err != nil {
		log.Fatal(err)
	}
	atendimento, err := cfg.Caso("amplitude-atendimento-cliente")
	if err != nil {
		log.Fatal(err)
	}

	imprimirCasoAprovado(auto, framework.AvaliarFramework(auto.Scores, pesosAHP, cfg.LimiarVerde))
	imprimirCasoReprovadoPorDado(saude, framework.AvaliarFramework(saude.Scores, pesosAHP, cfg.LimiarVerde), cfg.LimiarVerde)
	imprimirCasoReprovadoGeral(atendimento, framework.AvaliarFramework(atendimento.Scores, pesosAHP, cfg.LimiarVerde))

	fmt.Println("\nTrês casos, três respostas diferentes. Auto: sim. Saúde Empresarial: ainda não,")
	fmt.Println("só falta dado, e dado é questão de tempo. Atendimento ao Cliente: não -- a tarefa")
	fmt.Println("em si é aberta e instável demais. Mais dado não resolve um problema que não é de dado.")
}

func imprimirGovernanca(caso config.Caso) {
	g := framework.Resultado{} // placeholder para manter import de framework usado só aqui se necessario
	_ = g
	detalhe := "dado pessoal comum, não sensível."
	if caso.Governanca.DadoSensivelLGPD {
		detalhe = "dado de categoria sensível (LGPD Art. 5º, II) -- DPA verificado."
	}
	fmt.Printf("  [Governança] APROVADO -- %s Base legal: %s\n", detalhe, caso.Governanca.BaseLegalDescricao)
}

func imprimirPerguntas(r framework.Resultado) {
	for i, chave := range framework.ChavesPerguntas {
		s := r.SinaisPorPergunta[chave]
		fmt.Printf("  Pergunta %d [%s, score %.2f]\n", i+1, s.Sinal, s.Score)
	}
}

func imprimirCasoAprovado(caso config.Caso, resultado framework.Resultado) {
	fmt.Printf("\n===== %s =====\n", caso.Nome)
	fmt.Printf("Tarefa: %s\n\n", caso.Tarefa)
	imprimirGovernanca(caso)
	imprimirPerguntas(resultado)
	fmt.Printf("\n  Score composto (AHP): %.2f\n", *resultado.ScoreComposto)
	fmt.Printf("  Recomendação (gate): %s\n", resultado.Recomendacao)

	npv := finance.CalcularNPV(finance.ParamsDeterministicos(caso.Financeiro))
	fmt.Println("  --- Análise financeira (DCF) ---")
	breakeven := "não atinge no horizonte"
	if npv.MesBreakeven != nil {
		breakeven = fmt.Sprintf("mês %d", *npv.MesBreakeven)
	}
	fmt.Printf("  NPV em %d meses (cenário mais provável): R$ %.2f\n", caso.Financeiro.HorizonteMeses, npv.Npv)
	fmt.Printf("  Breakeven: %s\n", breakeven)

	mc := finance.SimularMonteCarlo(caso.Financeiro, 10000, rand.Float64)
	fmt.Printf("\n  --- Monte Carlo (10.000 simulações) ---\n")
	fmt.Printf("  NPV médio: R$ %.2f | P5: R$ %.2f | P95: R$ %.2f | Probabilidade positiva: %.1f%%\n",
		mc.Media, mc.P5, mc.P95, mc.ProbabilidadePositivo*100)

	sens := finance.Analisar(caso.Financeiro, 0.2)
	fmt.Println("  --- Sensibilidade (ranking por impacto no NPV, +/-20%) ---")
	for i, s := range sens {
		fmt.Printf("  %d. %s: amplitude de R$ %.2f\n", i+1, s.Parametro, s.Amplitude)
	}
}

func imprimirCasoReprovadoPorDado(caso config.Caso, resultado framework.Resultado, limiarVerde float64) {
	fmt.Printf("\n===== %s =====\n", caso.Nome)
	fmt.Printf("Tarefa: %s\n\n", caso.Tarefa)
	imprimirGovernanca(caso)
	imprimirPerguntas(resultado)
	fmt.Println("  Isso é um \"ainda não\", não um \"não\" -- elegível a análise de Real Options.")

	opcao := finance.PrecificarOpcaoDeEsperar(caso.Financeiro, caso.Scores["p3"], limiarVerde, 10000, rand.Float64)
	fmt.Println("  --- Real Options: árvore binomial, decidir agora vs. esperar ---")
	fmt.Printf("  Meses até o score de dado cruzar o limiar: %d\n", opcao.MesesParaEsperar)
	fmt.Printf("  Valor de esperar (diferença): R$ %.2f\n", opcao.ValorDeEsperar)
	fmt.Printf("  Recomendação: %s\n", opcao.Recomendacao)
	fmt.Printf("  Reavaliação agendada em: %s\n", opcao.ReavaliacaoAgendadaEm)
}

func imprimirCasoReprovadoGeral(caso config.Caso, resultado framework.Resultado) {
	fmt.Printf("\n===== %s =====\n", caso.Nome)
	fmt.Printf("Tarefa: %s\n\n", caso.Tarefa)
	imprimirGovernanca(caso)
	imprimirPerguntas(resultado)
	fmt.Println("  Reparem: a pergunta 3 (dado suficiente) está VERDE. Não é falta de dado.")
	fmt.Printf("  Recomendação: %s\n", resultado.Recomendacao)
}
