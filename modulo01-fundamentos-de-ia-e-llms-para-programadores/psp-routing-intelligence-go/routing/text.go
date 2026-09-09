// Package routing e a camada de aplicacao do pipeline RAG — equivalente ao
// pacote com.psprouting.application da versao Java, com a mesma divisao de
// responsabilidades (serializacao de texto, montagem do prompt, parsing da
// resposta do LLM, orquestracao do pipeline e do seed), porem organizadas
// em funcoes puras e um struct de servico com campos de funcao em vez de
// classes @Service — mais idiomatico em Go do que replicar a injecao de
// dependencia do Spring.
package routing

import (
	"fmt"

	"psp-routing-intelligence/domain"
)

// SerializeTransaction gera o texto natural em pt-BR usado para o embedding
// — passo (1) do pipeline, equivalente a
// TransactionTextSerializer.serialize(Transaction) da versao Java:
//
//	"Pagamento via CREDIT_CARD, valor R$1500.00, bandeira VISA,
//	 regiao SP, hora 20h, categoria STREAMING."
func SerializeTransaction(tx domain.Transaction) string {
	return serialize(tx.Amount, tx.Method, tx.Brand, tx.UserRegion, tx.Hour, tx.MerchantCategory)
}

// SerializeHistorical gera o mesmo texto natural para uma transacao
// historica do seed — equivalente a
// TransactionTextSerializer.serialize(HistoricalTransaction).
func SerializeHistorical(tx domain.HistoricalTransaction) string {
	return serialize(tx.Amount, tx.Method, tx.Brand, tx.Region, tx.Hour, tx.Category)
}

func serialize(amount float64, method domain.PaymentMethod, brand, region string, hour int, category string) string {
	text := fmt.Sprintf("Pagamento via %s, valor R$%.2f", method, amount)

	if brand != "" {
		text += fmt.Sprintf(", bandeira %s", brand)
	}

	text += fmt.Sprintf(", regiao %s, hora %dh, categoria %s.", region, hour, category)
	return text
}
