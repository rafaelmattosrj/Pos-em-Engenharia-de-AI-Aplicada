// Comando decision-framework-tool imprime a demo do Framework de Decisao de
// Tres Perguntas: os 4 casos do Bloco 1 e a decomposicao da tarefa hibrida de
// referencia (Bloco 2), mesmo espirito da versao JS/Python original — mostrar
// o mecanismo funcionando, nao so os testes passando silenciosamente.
//
// Diferenca deliberada em relacao ao original: la, os testes automatizados
// (node:assert / unittest) e a demo rodam na mesma execucao direta do arquivo,
// porque JS e Python nao separam "codigo de teste" de "codigo de producao" em
// comandos distintos por padrao. Em Go, a convencao do projeto (mesma dos
// demais portes do repositorio) e separar os testes automatizados (rodados via
// `go test ./...`) do binario de demo (`go run .`). O comportamento observavel
// da demo em si e identico ao original; so a orquestracao teste+demo-no-mesmo-
// comando nao foi replicada.
//
// Uso: go run .
package main

import (
	"fmt"
	"strings"

	"decision-framework-tool/decision"
)

func main() {
	linhaDupla := strings.Repeat("=", 72)

	fmt.Println(linhaDupla)
	fmt.Println("Bloco 1: as 4 combinações da árvore pura de três perguntas")
	fmt.Println(linhaDupla)

	casosBloco1 := []struct {
		rotulo     string
		p1, p2, p3 bool
	}{
		{"P1=True,  P2=False, P3=False", true, false, false},
		{"P1=False, P2=True,  P3=False", false, true, false},
		{"P1=False, P2=False, P3=True ", false, false, true},
		{"P1=False, P2=False, P3=False", false, false, false},
	}
	for _, caso := range casosBloco1 {
		classificacao := decision.ClassificarTarefa(caso.p1, caso.p2, caso.p3)
		fmt.Printf("  %s -> %s\n", caso.rotulo, classificacao)
	}

	fmt.Println()
	fmt.Println(linhaDupla)
	fmt.Println("Bloco 2: decompondo a tarefa híbrida de referência")
	fmt.Println("(TrialForge - Emenda de Protocolo)")
	fmt.Println(linhaDupla)

	subtarefas := []decision.Subtarefa{
		{
			Nome: "Extrair o que mudou entre versões do protocolo",
			Tipo: "Extração/Interpretação",
			P1:   false,
			P2:   false,
			P3:   true,
		},
		{
			Nome: "Classificar o tipo de emenda (administrativa/substancial)",
			Tipo: "Decisão de Negócio",
			P1:   true,
			P2:   false,
			P3:   false,
		},
	}

	for _, subtarefa := range decision.DecomporTarefaHibrida(subtarefas) {
		fmt.Printf("  Subtarefa: %s\n", subtarefa.Nome)
		fmt.Printf("    Tipo: %s\n", subtarefa.Tipo)
		fmt.Printf("    P1=%v, P2=%v, P3=%v\n", subtarefa.P1, subtarefa.P2, subtarefa.P3)
		fmt.Printf("    Classificação: %s\n", subtarefa.Classificacao)
		fmt.Println()
	}

	fmt.Println("  Nota: as outras duas linhas da tabela de referência do checklist")
	fmt.Println("  ('Rotear pela criticidade' e 'Regenerar documentos afetados') não")
	fmt.Println("  entram nesta demo de propósito, pelo mesmo motivo documentado em")
	fmt.Println("  classifier_test.go: misturam regra e gate condicional de um jeito")
	fmt.Println("  que não mapeia limpo pra uma única resposta P1/P2/P3.")
}
