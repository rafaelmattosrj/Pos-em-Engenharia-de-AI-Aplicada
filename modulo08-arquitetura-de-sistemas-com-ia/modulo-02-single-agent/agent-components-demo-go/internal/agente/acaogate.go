package agente

// AcaoProposta e uma acao que o agente quer executar.
//
// Adaptacao: no original, Executar e uma funcao/lambda sem argumentos
// embutida no objeto literal da acao proposta. Aqui vira um func() string
// dentro da struct — mesmo formato observavel (so roda se nao houver gate).
type AcaoProposta struct {
	Tipo            string
	RequerAprovacao bool
	Executar        func() string
}

// ResultadoAcao e o retorno de ExecutarOuGatear: ou foi "executada" (com
// Resultado preenchido), ou ficou "aguardando_aprovacao" (com Mensagem
// preenchida) — nunca os dois ao mesmo tempo, igual ao original.
type ResultadoAcao struct {
	Status    string
	Resultado string
	Mensagem  string
}

// ExecutarOuGatear e o Approval Gate: nem toda acao proposta executa direto.
func ExecutarOuGatear(acaoProposta AcaoProposta) ResultadoAcao {
	if acaoProposta.RequerAprovacao {
		return ResultadoAcao{
			Status: "aguardando_aprovacao",
			Mensagem: "Ação \"" + acaoProposta.Tipo + "\" NÃO executada. Aguardando Approval Gate " +
				"(Módulo 1.3, Pergunta 2: erro caro e irreversível).",
		}
	}
	return ResultadoAcao{
		Status:    "executada",
		Resultado: acaoProposta.Executar(),
	}
}
