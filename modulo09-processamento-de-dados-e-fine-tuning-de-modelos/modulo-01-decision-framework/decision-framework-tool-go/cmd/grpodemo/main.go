// Demo do GRPO (Group Relative Policy Optimization): amostragem real de
// grupo via Ollama local -> recompensa verificável -> vantagem relativa ao
// grupo. NÃO treina nada -- reproduz só o sinal de aprendizado. Equivalente
// ao main() de grpo-verifiable-reward-demo.js / GrpoDemoMain.java. Requer
// `ollama serve` rodando localmente com o modelo "gemma4:e2b" disponível.
package main

import (
	"fmt"

	"decision-framework-tool/grpo"
)

const (
	modelo       = "gemma4:e2b"
	tamanhoGrupo = 6
	temperatura  = 1.0
)

const documentoSinistro = `Amplitude Seguros - Comunicado de Sinistro (documento digitalizado, OCR, scan parcialmente ilegivel)
Ref. anterior (cancelada): AS-2Q25-0988zi
Apolice vigente: A5-2026-1l44/7 (numero de dificil leitura no scan)
Segurado: M4rcos Vin1cius Alm3ida Te1xeira (OCR com ruido no nome)
Tipo de sinistro: Colisao veicular (ou possivelmente "Colisao e furto", trecho cortado)
Data do incidente: 12/O7/2026 ou 17/02/2026 (data ambigua, dois carimbos sobrepostos)
Valor estimado do reparo: R$ 8.45O,OO (sujeito a pericia, valor preliminar)
Status atual: Em analise pericial
`

var promptSchema = fmt.Sprintf(`Extraia os dados do documento abaixo e responda APENAS com um JSON
valido, sem nenhum texto adicional, seguindo exatamente este schema:
{"claimant_name": string, "claim_type": string, "incident_date": "DD/MM/AAAA",
"estimated_amount_brl": number, "status": string,
"prioridade": "alta se estimated_amount_brl > 5000, senao baixa"}

Documento:
%s
`, documentoSinistro)

func main() {
	ollama := grpo.NewOllamaClient()

	fmt.Println("== Amostragem real de grupo (o passo que o GRPO chama de 'rollout') ==")
	fmt.Printf("Modelo: %s  |  tamanho do grupo G=%d  |  temperatura=%.1f\n", modelo, tamanhoGrupo, temperatura)

	if _, err := ollama.Chamar(modelo, "teste de conexao", 0.1); err != nil {
		fmt.Printf("Não foi possível conectar ao Ollama local: %s\n", err)
		fmt.Println("Rode 'ollama serve' e confirme que o modelo 'gemma4:e2b' está disponível ('ollama list').")
		return
	}

	var recompensas []float64
	for i := 0; i < tamanhoGrupo; i++ {
		bruto, err := ollama.Chamar(modelo, promptSchema, temperatura)
		if err != nil {
			fmt.Printf("  [%d] chamada ao Ollama falhou (%s) -- pulando esta amostra do grupo.\n", i, err)
			continue
		}
		candidato := grpo.ExtrairJSON(bruto)
		r := grpo.RecompensaVerificavel(candidato)
		recompensas = append(recompensas, r)
		fmt.Printf("  [%d] recompensa=%.2f  json_valido=%v\n", i, r, candidato != nil)
	}

	if len(recompensas) == 0 {
		fmt.Println("Nenhuma amostra completou -- sem grupo pra calcular vantagem.")
		return
	}

	resultado := grpo.VantagemRelativaAoGrupo(recompensas)

	fmt.Printf("\nGrupo completo: media(r)=%.3f  desvio_padrao(r)=%.3f\n", resultado.Media, resultado.Desvio)

	if resultado.Desvio == 0.0 {
		fmt.Println("GRUPO DEGENERADO: todas as G respostas receberam a mesma recompensa -- sinal de aprendizado desaparece nessa rodada.")
	} else {
		for i, r := range recompensas {
			a := resultado.Vantagens[i]
			var tag string
			switch {
			case a > 0:
				tag = "reforçaria essa resposta (A>0)"
			case a < 0:
				tag = "penalizaria essa resposta (A<0)"
			default:
				tag = "neutro (A=0)"
			}
			fmt.Printf("%2d  recompensa=%.2f  A_i=%.3f  %s\n", i, r, a, tag)
		}
	}

	fmt.Println("\nNenhum peso do modelo foi atualizado por este script -- fim do demo.")
}
