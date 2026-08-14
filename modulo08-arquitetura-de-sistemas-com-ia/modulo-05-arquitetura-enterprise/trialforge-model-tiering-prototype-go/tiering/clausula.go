// Package tiering implementa a cascata de Model Tiering (FrugalGPT) com
// orçamento por estudo, RAG de cláusulas, Approval Gate e trilha de
// auditoria. Porte de trialforge-model-tiering-prototype.js / .py.
package tiering

// Clausula é uma cláusula do banco do Agente ICF: tema (usado na busca),
// texto e fonte regulatória.
type Clausula struct {
	Tema  string
	Texto string
	Fonte string
}

// BancoClausulas: mesmo dado do Módulo 2.5/4.5/5.4.
var BancoClausulas = []Clausula{
	{
		Tema: "Assentimento de menores de idade em estudos clínicos",
		Texto: "Para participantes entre 12 e 17 anos, é necessário assentimento por escrito, " +
			"além do consentimento do responsável legal (RDC ANVISA 466/2012, Art. 4º).",
		Fonte: "RDC ANVISA 466/2012, Art. 4º",
	},
	{
		Tema: "Direito de retirada do participante do estudo a qualquer momento",
		Texto: "O participante pode retirar seu consentimento a qualquer momento, sem necessidade " +
			"de justificativa e sem prejuízo ao seu tratamento (RDC ANVISA 466/2012, Art. 5º).",
		Fonte: "RDC ANVISA 466/2012, Art. 5º",
	},
}
