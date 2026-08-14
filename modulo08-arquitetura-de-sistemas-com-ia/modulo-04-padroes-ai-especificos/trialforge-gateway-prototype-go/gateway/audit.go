package gateway

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// AuditTrail (Módulo 4.4 — Audit Trail) é a trilha de auditoria append-only:
// nunca sobrescreve (21 CFR Part 11) — cada registro é uma linha JSON
// acrescentada ao arquivo.
type AuditTrail struct {
	Caminho string
}

// NewAuditTrail cria uma trilha de auditoria apontando pro arquivo indicado.
func NewAuditTrail(caminho string) *AuditTrail {
	return &AuditTrail{Caminho: caminho}
}

// Registrar acrescenta um registro (com timestamp automático) ao arquivo de
// auditoria. `registro` usa chaves em snake_case pra bater com o formato dos
// dois originais (id_requisicao, cache_hit, gate_acionado, etc.).
func (t *AuditTrail) Registrar(registro map[string]interface{}) error {
	completo := make(map[string]interface{}, len(registro)+1)
	completo["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	for k, v := range registro {
		completo[k] = v
	}

	linha, err := json.Marshal(completo)
	if err != nil {
		return fmt.Errorf("audit: falha ao serializar registro: %w", err)
	}

	arquivo, err := os.OpenFile(t.Caminho, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("audit: falha ao abrir trilha: %w", err)
	}
	defer arquivo.Close()

	if _, err := arquivo.Write(append(linha, '\n')); err != nil {
		return fmt.Errorf("audit: falha ao gravar registro: %w", err)
	}
	return nil
}

func lerLinhasAuditoria(caminho string) ([]map[string]interface{}, error) {
	arquivo, err := os.Open(caminho)
	if err != nil {
		return nil, fmt.Errorf("audit: falha ao abrir trilha: %w", err)
	}
	defer arquivo.Close()

	var linhas []map[string]interface{}
	scanner := bufio.NewScanner(arquivo)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		texto := scanner.Text()
		if texto == "" {
			continue
		}
		var registro map[string]interface{}
		if err := json.Unmarshal([]byte(texto), &registro); err != nil {
			return nil, fmt.Errorf("audit: linha inválida na trilha: %w", err)
		}
		linhas = append(linhas, registro)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("audit: falha ao ler trilha: %w", err)
	}
	return linhas, nil
}

// campoStr/campoBool/campoFloat leem um campo de uma linha da trilha (que é
// um map[string]interface{} vindo de json.Unmarshal), devolvendo o zero-value
// quando ausente ou de tipo diferente — equivalente ao `campo(indice, chave)`
// do Python que devolve None fora dos limites.
func campoBool(linha map[string]interface{}, chave string) (bool, bool) {
	v, ok := linha[chave]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func campoStr(linha map[string]interface{}, chave string) (string, bool) {
	v, ok := linha[chave]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func campoFloat(linha map[string]interface{}, chave string) (float64, bool) {
	v, ok := linha[chave]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

// checagem é uma verificação nomeada da trilha de auditoria.
type checagem struct {
	Descricao string
	OK        bool
}

// VerificarTrilhaAuditoria relê o audit-trail.jsonl e confere que as ÚLTIMAS
// 5 requisições CONCLUÍDAS (ignorando pendências "aguardando_aprovacao")
// tomaram as decisões determinísticas esperadas pelo roteiro de demo (ver
// main.go) — cache hit/miss, modelo escolhido, índice roteado, gate acionado
// ou não. Nunca compara o texto exato gerado pelo modelo, que varia entre
// execuções. Não confia só no humano assistindo notar os comportamentos.
func VerificarTrilhaAuditoria(caminho string, modeloCaro string, limiarConfianca float64, logf func(format string, args ...interface{})) error {
	todasLinhas, err := lerLinhasAuditoria(caminho)
	if err != nil {
		return err
	}

	var linhasConcluidas []map[string]interface{}
	for _, l := range todasLinhas {
		status, _ := campoStr(l, "status_final")
		if status != "aguardando_aprovacao" {
			linhasConcluidas = append(linhasConcluidas, l)
		}
	}

	inicio := len(linhasConcluidas) - 5
	if inicio < 0 {
		inicio = 0
	}
	linhas := linhasConcluidas[inicio:]

	campo := func(indice int, chave string) (map[string]interface{}, bool) {
		if indice < 0 || indice >= len(linhas) {
			return nil, false
		}
		return linhas[indice], true
	}
	cB := func(indice int, chave string) (bool, bool) {
		l, ok := campo(indice, chave)
		if !ok {
			return false, false
		}
		return campoBool(l, chave)
	}
	cS := func(indice int, chave string) (string, bool) {
		l, ok := campo(indice, chave)
		if !ok {
			return "", false
		}
		return campoStr(l, chave)
	}
	cF := func(indice int, chave string) (float64, bool) {
		l, ok := campo(indice, chave)
		if !ok {
			return 0, false
		}
		return campoFloat(l, chave)
	}

	if logf != nil {
		logf("\n== Verificação: últimas 5 requisições concluídas (%d concluídas, %d entradas no total incluindo pendências) ==", len(linhasConcluidas), len(todasLinhas))
	}

	algumaPendencia := false
	for _, l := range todasLinhas {
		if s, _ := campoStr(l, "status_final"); s == "aguardando_aprovacao" {
			algumaPendencia = true
			break
		}
	}

	cacheHit0, _ := cB(0, "cache_hit")
	gate0, _ := cB(0, "gate_acionado")
	cacheHit1, _ := cB(1, "cache_hit")
	status1, _ := cS(1, "status_final")
	modelo2, _ := cS(2, "modelo_usado")
	gate2, _ := cB(2, "gate_acionado")
	indice2, _ := cS(2, "indice_usado")
	cacheHit3, _ := cB(3, "cache_hit")
	iteracoes3, iteracoes3ok := cF(3, "iteracoes_agentic")
	esgotou3, _ := cB(3, "esgotou_agentic")
	gate3, _ := cB(3, "gate_acionado")
	confianca3, confianca3ok := cF(3, "confianca_rag")
	indice4, _ := cS(4, "indice_usado")
	iteracoes4, _ := cF(4, "iteracoes_agentic")
	gate4, _ := cB(4, "gate_acionado")

	checagens := []checagem{
		{"pelo menos 5 requisições concluídas na trilha", len(linhasConcluidas) >= 5},
		{`ao menos 1 registro "aguardando_aprovacao" gravado antes de qualquer decisão`, algumaPendencia},
		{"#1 rotina: sem cache hit", !cacheHit0},
		{"#1 rotina: confiança alta, sem Approval Gate", !gate0},
		{"#2 paráfrase: cache HIT", cacheHit1 && status1 == "respondido_via_cache"},
		{"#3 síntese de CSR: modelo caro", modelo2 == modeloCaro},
		{"#3 síntese de CSR: Approval Gate sempre acionado", gate2},
		{"#3 síntese de CSR: Multi-Index roteou pro índice csr", indice2 == "csr"},
		{"#4 tema diferente: cache miss", !cacheHit3},
		{"#4 tema diferente: Agentic RAG esgotou as 3 iterações", iteracoes3ok && int(iteracoes3) == MaxIteracoesAgentic && esgotou3},
		{"#4 tema diferente: Approval Gate por confiança baixa", gate3 && confianca3ok && confianca3 < limiarConfianca},
		{"#5 critério de protocolo: Multi-Index roteou pro índice protocolo", indice4 == "protocolo"},
		{"#5 critério de protocolo: Agentic RAG confiante já na 1ª iteração", int(iteracoes4) == 1},
		{"#5 critério de protocolo: sem Approval Gate", !gate4},
	}

	passou := 0
	for _, c := range checagens {
		if logf != nil {
			status := "FALHOU"
			if c.OK {
				status = "OK"
			}
			logf("  [%s] %s", status, c.Descricao)
		}
		if c.OK {
			passou++
		}
	}
	if logf != nil {
		logf("Total: %d verificação(ões), %d passou(passaram), %d falhou(falharam).", len(checagens), passou, len(checagens)-passou)
	}

	if passou != len(checagens) {
		return fmt.Errorf("a trilha de auditoria não confirma os 4 caminhos esperados (%d/%d) — reveja audit-trail.jsonl", passou, len(checagens))
	}
	return nil
}
