// Package domain contem os tipos de negocio do PSP Routing Intelligence,
// portados 1:1 do pacote com.psprouting.domain da versao Java: os mesmos
// enums (PSP, PaymentMethod, TransactionStatus) e os mesmos registros
// (Transaction, HistoricalTransaction, SimilarCase, RoutingRecommendation).
package domain

import (
	"fmt"
	"strings"
)

// PSP e um Payment Service Provider suportado pelo roteamento — mesmo
// conjunto fixo de com.psprouting.domain.PSP.
type PSP string

const (
	Adyen       PSP = "ADYEN"
	Braspag     PSP = "BRASPAG"
	Bradesco    PSP = "BRADESCO"
	Santander   PSP = "SANTANDER"
	Nupay       PSP = "NUPAY"
	Picpay      PSP = "PICPAY"
	MercadoPago PSP = "MERCADOPAGO"
)

// ParsePSP valida e normaliza uma string como PSP, retornando erro para
// valores fora do enum — usado tanto ao decodificar a resposta do LLM
// quanto os metadados lidos do Neo4j.
func ParsePSP(value string) (PSP, error) {
	switch psp := PSP(strings.ToUpper(strings.TrimSpace(value))); psp {
	case Adyen, Braspag, Bradesco, Santander, Nupay, Picpay, MercadoPago:
		return psp, nil
	default:
		return "", fmt.Errorf("PSP desconhecido: %q", value)
	}
}

// PaymentMethod e o metodo de pagamento de uma transacao.
type PaymentMethod string

const (
	Pix        PaymentMethod = "PIX"
	CreditCard PaymentMethod = "CREDIT_CARD"
	Wallet     PaymentMethod = "WALLET"
)

// ParsePaymentMethod valida e normaliza uma string como PaymentMethod.
func ParsePaymentMethod(value string) (PaymentMethod, error) {
	switch method := PaymentMethod(strings.ToUpper(strings.TrimSpace(value))); method {
	case Pix, CreditCard, Wallet:
		return method, nil
	default:
		return "", fmt.Errorf("metodo de pagamento desconhecido: %q", value)
	}
}

// TransactionStatus e o desfecho de uma transacao historica.
type TransactionStatus string

const (
	Success TransactionStatus = "SUCCESS"
	Failed  TransactionStatus = "FAILED"
)

// ParseTransactionStatus valida e normaliza uma string como TransactionStatus.
func ParseTransactionStatus(value string) (TransactionStatus, error) {
	switch status := TransactionStatus(strings.ToUpper(strings.TrimSpace(value))); status {
	case Success, Failed:
		return status, nil
	default:
		return "", fmt.Errorf("status de transacao desconhecido: %q", value)
	}
}

// Transaction e a transacao recebida em POST /api/routing/recommend — as
// caracteristicas usadas para gerar o embedding e buscar casos similares.
//
// Brand e opcional: PIX e WALLET normalmente nao tem bandeira de cartao.
type Transaction struct {
	Amount           float64
	Method           PaymentMethod
	Brand            string
	UserRegion       string
	Hour             int
	MerchantCategory string
}

// HistoricalTransaction e uma transacao de data/transactions-seed.json,
// usada para popular o Neo4j com casos reais de aprovacao/reprovacao por
// PSP. Mesmo shape de Transaction, acrescido do PSP que processou a
// transacao e do status do desfecho.
type HistoricalTransaction struct {
	Amount   float64           `json:"amount"`
	Method   PaymentMethod     `json:"method"`
	Brand    string            `json:"brand"`
	Region   string            `json:"region"`
	Hour     int               `json:"hour"`
	Category string            `json:"category"`
	PSP      PSP               `json:"psp"`
	Status   TransactionStatus `json:"status"`
}

// SimilarCase e um caso historico recuperado do Neo4j por similaridade de
// cosseno com a transacao atual.
type SimilarCase struct {
	PSP        PSP               `json:"psp"`
	Status     TransactionStatus `json:"status"`
	Amount     float64           `json:"amount"`
	Method     PaymentMethod     `json:"method"`
	Similarity float64           `json:"similarity"`
}

// RoutingRecommendation e o resultado completo do pipeline de roteamento —
// corpo de resposta de POST /api/routing/recommend, no mesmo shape da
// versao Java (com.psprouting.domain.RoutingRecommendation).
type RoutingRecommendation struct {
	Primary      PSP           `json:"primary"`
	Confidence   float64       `json:"confidence"`
	Reasoning    string        `json:"reasoning"`
	Fallback     []PSP         `json:"fallback"`
	SimilarCases []SimilarCase `json:"similarCases"`
}
