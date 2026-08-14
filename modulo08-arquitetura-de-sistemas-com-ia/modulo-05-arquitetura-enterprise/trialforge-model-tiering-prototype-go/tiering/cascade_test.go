package tiering

import (
	"context"
	"path/filepath"
	"testing"
)

func construirIndice(t *testing.T, fake *fakeGateway) *RagIndex {
	t.Helper()
	fake.comEmbedding(BancoClausulas[0].Tema, []float64{1, 0}).
		comEmbedding(BancoClausulas[0].Texto, []float64{1, 0}).
		comEmbedding(BancoClausulas[1].Tema, []float64{0, 1}).
		comEmbedding(BancoClausulas[1].Texto, []float64{0, 1})
	ragIndex := NewRagIndex(fake, ModeloEmbedding)
	if err := ragIndex.PrepararIndice(context.Background()); err != nil {
		t.Fatal(err)
	}
	return ragIndex
}

// ---------- Cenário 1: Tier 1 resolve sozinho, sem escalar ----------

func TestProcessarComCascata_ConfiancaAlta_ResolveNoTier1SemEscalar(t *testing.T) {
	pergunta := "pergunta-rotina-tier1"
	fake := newFakeGateway().
		comEmbedding(pergunta, []float64{1, 0}).
		comEmbedding("resposta-tier1-rotina", []float64{1, 0}).
		comChat(ModeloTier1, pergunta, "resposta-tier1-rotina")
	ragIndex := construirIndice(t, fake)
	orcamento := NewOrcamentoManager()
	orcamento.DefinirOrcamento("estudo-A", 0.05)
	auditTrail := NewAuditTrail(filepath.Join(t.TempDir(), "audit.jsonl"))
	gateway := NewCascadeGateway(fake, ragIndex, orcamento, newFakeApprovalPrompt(true), auditTrail)

	rascunho, err := gateway.ProcessarComCascata(context.Background(), pergunta, "estudo-A")
	if err != nil {
		t.Fatal(err)
	}
	if rascunho != "resposta-tier1-rotina" {
		t.Errorf("esperava \"resposta-tier1-rotina\", obteve %q", rascunho)
	}
	if !closeEnough(orcamento.GastoAtual("estudo-A"), CustoTier1) {
		t.Errorf("esperava gasto %v, obteve %v", CustoTier1, orcamento.GastoAtual("estudo-A"))
	}

	linhas, _ := auditTrail.LerTodas()
	if len(linhas) != 1 {
		t.Fatalf("esperava 1 linha de auditoria, obteve %d", len(linhas))
	}
	if linhas[0]["tier_usado"] != "Tier 1" {
		t.Errorf("esperava tier_usado=Tier 1, obteve %v", linhas[0]["tier_usado"])
	}
	if linhas[0]["escalou_cascata"] != false {
		t.Errorf("esperava escalou_cascata=false, obteve %v", linhas[0]["escalou_cascata"])
	}
	if linhas[0]["status_final"] != "aprovado" {
		t.Errorf("esperava status_final=aprovado, obteve %v", linhas[0]["status_final"])
	}
}

// ---------- Cenário 2: pergunta fora do domínio -> escala pro Tier 2 ----------

func TestProcessarComCascata_ConfiancaBaixaDeBusca_EscalaParaTier2(t *testing.T) {
	pergunta := "pergunta-fora-dominio-escalada"
	fake := newFakeGateway().
		comEmbedding(pergunta, []float64{1, 1}). // equidistante dos dois temas, abaixo do limiar
		comEmbedding("resposta-tier1-fora-dominio", []float64{1, 1}).
		comEmbedding("resposta-tier2-escalada", []float64{1, 0}).
		comChat(ModeloTier1, pergunta, "resposta-tier1-fora-dominio").
		comChat(ModeloTier2, pergunta, "resposta-tier2-escalada")
	ragIndex := construirIndice(t, fake)
	orcamento := NewOrcamentoManager()
	orcamento.DefinirOrcamento("estudo-A", 0.05)
	auditTrail := NewAuditTrail(filepath.Join(t.TempDir(), "audit.jsonl"))
	gateway := NewCascadeGateway(fake, ragIndex, orcamento, newFakeApprovalPrompt(true), auditTrail)

	rascunho, err := gateway.ProcessarComCascata(context.Background(), pergunta, "estudo-A")
	if err != nil {
		t.Fatal(err)
	}
	if rascunho != "resposta-tier2-escalada" {
		t.Errorf("esperava \"resposta-tier2-escalada\", obteve %q", rascunho)
	}
	if !closeEnough(orcamento.GastoAtual("estudo-A"), CustoTier1+CustoTier2) {
		t.Errorf("esperava gasto %v, obteve %v", CustoTier1+CustoTier2, orcamento.GastoAtual("estudo-A"))
	}

	linhas, _ := auditTrail.LerTodas()
	if linhas[0]["tier_usado"] != "Tier 2 (escalado)" {
		t.Errorf("esperava tier_usado=Tier 2 (escalado), obteve %v", linhas[0]["tier_usado"])
	}
	if linhas[0]["escalou_cascata"] != true {
		t.Errorf("esperava escalou_cascata=true, obteve %v", linhas[0]["escalou_cascata"])
	}
	if _, ok := linhas[0]["confianca_resposta"].(float64); !ok {
		t.Errorf("esperava confianca_resposta numérico, obteve %v (%T)", linhas[0]["confianca_resposta"], linhas[0]["confianca_resposta"])
	}
}

// ---------- Cenário 3: síntese de CSR -> regra fixa pro Tier 2 + Approval Gate ----------

func TestProcessarComCascata_SinteseDeCsr_VaiDireitoParaTier2ComAprovacao(t *testing.T) {
	pergunta := "Preciso do CSR final agora, por favor."
	fake := newFakeGateway().
		comEmbedding(pergunta, []float64{1, 0}).
		comChat(ModeloTier2, "CSR", "resposta-tier2-csr")
	ragIndex := construirIndice(t, fake)
	orcamento := NewOrcamentoManager()
	orcamento.DefinirOrcamento("estudo-A", 0.05)
	auditTrail := NewAuditTrail(filepath.Join(t.TempDir(), "audit.jsonl"))
	aprovacao := newFakeApprovalPrompt(true)
	gateway := NewCascadeGateway(fake, ragIndex, orcamento, aprovacao, auditTrail)

	rascunho, err := gateway.ProcessarComCascata(context.Background(), pergunta, "estudo-A")
	if err != nil {
		t.Fatal(err)
	}
	if rascunho != "resposta-tier2-csr" {
		t.Errorf("esperava \"resposta-tier2-csr\", obteve %q", rascunho)
	}
	if aprovacao.chamadas != 1 {
		t.Errorf("esperava 1 chamada de aprovação, obteve %d", aprovacao.chamadas)
	}
	if aprovacao.ultimoRascunho != "resposta-tier2-csr" {
		t.Errorf("esperava rascunho \"resposta-tier2-csr\" no Approval Gate, obteve %q", aprovacao.ultimoRascunho)
	}
	// regra fixa: só Tier 2 é chamado, sem cascata (nunca tenta Tier 1 primeiro)
	if !closeEnough(orcamento.GastoAtual("estudo-A"), CustoTier2) {
		t.Errorf("esperava gasto %v, obteve %v", CustoTier2, orcamento.GastoAtual("estudo-A"))
	}

	linhas, _ := auditTrail.LerTodas()
	if linhas[0]["tier_usado"] != "Tier 2" {
		t.Errorf("esperava tier_usado=Tier 2, obteve %v", linhas[0]["tier_usado"])
	}
	if linhas[0]["escalou_cascata"] != false {
		t.Errorf("esperava escalou_cascata=false, obteve %v", linhas[0]["escalou_cascata"])
	}
	if linhas[0]["aprovado"] != true {
		t.Errorf("esperava aprovado=true, obteve %v", linhas[0]["aprovado"])
	}
	if linhas[0]["confianca_resposta"] != nil {
		t.Errorf("esperava confianca_resposta nulo, obteve %v", linhas[0]["confianca_resposta"])
	}
}

func TestProcessarComCascata_SinteseDeCsrReprovada_MarcaRejeitado(t *testing.T) {
	pergunta := "Preciso do CSR final agora, por favor."
	fake := newFakeGateway().
		comEmbedding(pergunta, []float64{1, 0}).
		comChat(ModeloTier2, "CSR", "resposta-tier2-csr")
	ragIndex := construirIndice(t, fake)
	orcamento := NewOrcamentoManager()
	orcamento.DefinirOrcamento("estudo-A", 0.05)
	auditTrail := NewAuditTrail(filepath.Join(t.TempDir(), "audit.jsonl"))
	gateway := NewCascadeGateway(fake, ragIndex, orcamento, newFakeApprovalPrompt(false), auditTrail)

	if _, err := gateway.ProcessarComCascata(context.Background(), pergunta, "estudo-A"); err != nil {
		t.Fatal(err)
	}

	linhas, _ := auditTrail.LerTodas()
	if linhas[0]["aprovado"] != false {
		t.Errorf("esperava aprovado=false, obteve %v", linhas[0]["aprovado"])
	}
	if linhas[0]["status_final"] != "rejeitado" {
		t.Errorf("esperava status_final=rejeitado, obteve %v", linhas[0]["status_final"])
	}
}

// ---------- Cenário 4: estudo com orçamento baixo -> bloqueia antes de chamar o modelo ----------

func TestProcessarComCascata_OrcamentoInsuficiente_BloqueiaAntesDeChamarModelo(t *testing.T) {
	pergunta := "pergunta-orcamento-baixo"
	fake := newFakeGateway() // nenhum chat/embed configurado: se for chamado, cai no fallback (não falha, mas não deveria ser chamado)
	ragIndex := construirIndice(t, fake)
	orcamento := NewOrcamentoManager()
	orcamento.DefinirOrcamento("estudo-B", 0.005) // menor que CustoTier1+CustoTier2 = 0.011
	auditTrail := NewAuditTrail(filepath.Join(t.TempDir(), "audit.jsonl"))
	gateway := NewCascadeGateway(fake, ragIndex, orcamento, newFakeApprovalPrompt(true), auditTrail)

	rascunho, err := gateway.ProcessarComCascata(context.Background(), pergunta, "estudo-B")
	if err != nil {
		t.Fatal(err)
	}
	if rascunho != "" {
		t.Errorf("esperava rascunho vazio (bloqueado), obteve %q", rascunho)
	}
	if orcamento.GastoAtual("estudo-B") != 0 {
		t.Errorf("esperava gasto 0, obteve %v", orcamento.GastoAtual("estudo-B"))
	}

	linhas, _ := auditTrail.LerTodas()
	if len(linhas) != 1 {
		t.Fatalf("esperava 1 linha de auditoria, obteve %d", len(linhas))
	}
	if linhas[0]["status_final"] != "bloqueado_por_orcamento" {
		t.Errorf("esperava status_final=bloqueado_por_orcamento, obteve %v", linhas[0]["status_final"])
	}
}

// ---------- Os 4 cenários em sequência, como em main(): a trilha de auditoria confirma tudo ----------

func TestQuatroCenariosEmSequencia_TrilhaDeAuditoriaConfirmaOsTresComportamentos(t *testing.T) {
	perguntaRotina := "pergunta-rotina-tier1"
	perguntaForaDominio := "pergunta-fora-dominio-escalada"
	perguntaCsr := "Preciso do CSR final agora, por favor."
	perguntaBloqueada := "pergunta-orcamento-baixo"

	fake := newFakeGateway().
		comEmbedding(perguntaRotina, []float64{1, 0}).
		comEmbedding("resposta-tier1-rotina", []float64{1, 0}).
		comChat(ModeloTier1, perguntaRotina, "resposta-tier1-rotina").
		comEmbedding(perguntaForaDominio, []float64{1, 1}).
		comEmbedding("resposta-tier1-fora-dominio", []float64{1, 1}).
		comEmbedding("resposta-tier2-escalada", []float64{1, 0}).
		comChat(ModeloTier1, perguntaForaDominio, "resposta-tier1-fora-dominio").
		comChat(ModeloTier2, perguntaForaDominio, "resposta-tier2-escalada").
		comEmbedding(perguntaCsr, []float64{1, 0}).
		comChat(ModeloTier2, "CSR", "resposta-tier2-csr")

	ragIndex := construirIndice(t, fake)
	orcamento := NewOrcamentoManager()
	orcamento.DefinirOrcamento("estudo-A", 0.05)
	orcamento.DefinirOrcamento("estudo-B", 0.005)
	auditTrail := NewAuditTrail(filepath.Join(t.TempDir(), "audit.jsonl"))
	gateway := NewCascadeGateway(fake, ragIndex, orcamento, newFakeApprovalPrompt(true), auditTrail)

	ctx := context.Background()
	if _, err := gateway.ProcessarComCascata(ctx, perguntaRotina, "estudo-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := gateway.ProcessarComCascata(ctx, perguntaForaDominio, "estudo-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := gateway.ProcessarComCascata(ctx, perguntaCsr, "estudo-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := gateway.ProcessarComCascata(ctx, perguntaBloqueada, "estudo-B"); err != nil {
		t.Fatal(err)
	}

	if err := VerificarEImprimir(auditTrail); err != nil {
		t.Fatalf("verificação da trilha de auditoria falhou: %v", err)
	}
}

// ---------- Volume concorrente: réplica de SimularVolumeConcorrente, com fake gateway ----------

func TestVolumeConcorrente_ReservarOrcamentoSeguraOOrcamentoPorEstudo(t *testing.T) {
	perguntaRotina := "pergunta-rotina-tier1"
	fake := newFakeGateway().
		comEmbeddingPadrao([]float64{1, 0}).
		comChatPadrao("resposta-padrao")
	ragIndex := construirIndice(t, fake)

	orcamento := NewOrcamentoManager()
	custoPiorCaso := CustoTier1 + CustoTier2
	limiteApertado := 2*custoPiorCaso + 0.0005 // cabe exatamente 2 reservas
	orcamento.DefinirOrcamento("estudo-F", limiteApertado)

	auditTrail := NewAuditTrail(filepath.Join(t.TempDir(), "audit.jsonl"))
	gateway := NewCascadeGateway(fake, ragIndex, orcamento, newFakeApprovalPrompt(true), auditTrail)

	ctx := context.Background()
	var requisicoes []func()
	for i := 0; i < 5; i++ {
		requisicoes = append(requisicoes, func() { gateway.ProcessarComCascata(ctx, perguntaRotina, "estudo-F") })
	}

	resultados := SimularVolumeConcorrente(orcamento, []string{"estudo-F"}, requisicoes)

	if len(resultados) != 1 {
		t.Fatalf("esperava 1 resultado, obteve %d", len(resultados))
	}
	if resultados[0].Estourou {
		t.Errorf("esperava não estourar, gasto %v limite %v", resultados[0].Gasto, resultados[0].Limite)
	}
	if orcamento.GastoAtual("estudo-F") > limiteApertado+1e-9 {
		t.Errorf("orçamento estourou: gasto %v, limite %v", orcamento.GastoAtual("estudo-F"), limiteApertado)
	}
}
