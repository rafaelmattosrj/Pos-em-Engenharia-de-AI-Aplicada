package tiering

import "fmt"

// Checagem é uma verificação individual sobre a trilha de auditoria.
type Checagem struct {
	Descricao string
	OK        bool
}

// VerificarTrilha relê a trilha de auditoria e confere as DECISÕES
// determinísticas (tier usado, se escalou, se bloqueou por orçamento), nunca
// o texto exato gerado pelo modelo. Olha só as últimas 4 entradas — a trilha
// é append-only, nunca apagada entre execuções.
//
// Porte 1:1 de verificarTrilhaAuditoria em trialforge-model-tiering-prototype.js / .py.
func VerificarTrilha(todasLinhas []map[string]any) []Checagem {
	linhas := todasLinhas
	if len(linhas) > 4 {
		linhas = linhas[len(linhas)-4:]
	}

	campo := func(indice int, chave string) any {
		if indice >= len(linhas) {
			return nil
		}
		return linhas[indice][chave]
	}
	texto := func(indice int, chave string) string {
		v, _ := campo(indice, chave).(string)
		return v
	}
	booleano := func(indice int, chave string) bool {
		v, _ := campo(indice, chave).(bool)
		return v
	}
	ehNumero := func(indice int, chave string) bool {
		_, ok := campo(indice, chave).(float64)
		return ok
	}

	return []Checagem{
		{"pelo menos 4 entradas na trilha", len(todasLinhas) >= 4},
		{"#1 estudo-A rotina: Tier 1 resolve, sem escalar",
			texto(0, "tier_usado") == "Tier 1" && !booleano(0, "escalou_cascata")},
		{"#2 estudo-A tema diferente: escala pro Tier 2",
			booleano(1, "escalou_cascata") && texto(1, "tier_usado") == "Tier 2 (escalado)"},
		{"#3 estudo-A síntese de CSR: regra fixa pro Tier 2, sem cascata",
			texto(2, "tier_usado") == "Tier 2" && !booleano(2, "escalou_cascata")},
		{"#3 síntese de CSR: aprovado no Approval Gate", booleano(2, "aprovado")},
		{"#4 estudo-B: bloqueado por orçamento antes de chamar o modelo",
			texto(3, "status_final") == "bloqueado_por_orcamento"},
		{"#1 e #2: escalação decidida pela confiança da RESPOSTA (g(pergunta,resposta)), não só da busca",
			ehNumero(0, "confianca_resposta") && ehNumero(1, "confianca_resposta")},
	}
}

// VerificarEImprimir executa VerificarTrilha sobre o conteúdo atual de
// auditTrail, imprime cada checagem e devolve erro se alguma falhar.
func VerificarEImprimir(auditTrail *AuditTrail) error {
	todasLinhas, err := auditTrail.LerTodas()
	if err != nil {
		return err
	}
	fmt.Printf("\n== Verificação: últimas 4 entradas da trilha de auditoria (%d no total) ==\n", len(todasLinhas))

	checagens := VerificarTrilha(todasLinhas)
	passou := 0
	for _, c := range checagens {
		status := "FALHOU"
		if c.OK {
			status = "OK"
			passou++
		}
		fmt.Printf("  [%s] %s\n", status, c.Descricao)
	}
	fmt.Printf("Total: %d verificação(ões), %d passou(passaram), %d falhou(falharam).\n",
		len(checagens), passou, len(checagens)-passou)

	if passou != len(checagens) {
		return fmt.Errorf("a trilha de auditoria não confirma os 3 comportamentos exigidos (%d/%d) — reveja audit-trail-tiering.jsonl",
			passou, len(checagens))
	}
	return nil
}
