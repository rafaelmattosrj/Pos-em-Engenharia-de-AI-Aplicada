---
name: resumo-explicacao-modulo
description: Gera os dois documentos padrão de estudo (Resumo e Explicação Detalhada) em .docx a partir do PDF de um módulo da Pós-Graduação em Engenharia de Software com IA Aplicada (UNIPDS). Use quando o usuário pedir para "criar o resumo e a explicação do módulo X", "documentar esse PDF da pós", "gerar os docs de estudo de [nome do módulo]", ou apontar um PDF de aula novo que ainda não tem os arquivos "- Resumo.docx" / "- Explicação.docx" correspondentes na raiz do repositório.
---

# Resumo + Explicação de Módulo

Gera, a partir de um único PDF de módulo (livro/apostila da aula), dois arquivos **.docx** na raiz do repositório (e uma cópia em `C:\Users\ACER\Documents\Pos em Engenharia de AI Aplicada`, que espelha a raiz do repo):

- `{Nome do Módulo} - Resumo.docx`
- `{Nome do Módulo} - Explicação.docx`

> ⚠️ **O entregável final é sempre `.docx`, nunca `.md`.** Já aconteceu de uma sessão anterior gerar 5 pares em `.md` por engano (achando que markdown puro era o padrão) — foram convertidos depois, mas o retrabalho era evitável. O Markdown é só um **rascunho intermediário em memória/scratchpad**, nunca commitado sozinho na raiz do repo.

Os dois seguem padrões **fixos e distintos**, validados nos módulos já produzidos: `Criação de Agentes Autônomos`, `Model Context Protocol (MCP)`, `Ferramentas de IA para DevOps`, `Ferramentas de IA para Gestão de projetos`, `Ferramentas de IA para UX & UI` (todos em `.docx` na raiz). Sempre abra um desses `.docx` como referência de tom, profundidade e formatação antes de escrever — eles são a fonte de verdade, mais do que qualquer descrição abaixo. Para ler o conteúdo de um `.docx` existente, use `python-docx` (já instalado) em vez de tentar abrir o binário diretamente.

## Passo 0 — Descobrir o que falta

Na raiz do repo, todo PDF de módulo deveria ter um par `- Resumo.docx` / `- Explicação.docx`. Liste os PDFs da raiz e cheque quais ainda não têm os dois arquivos. Nunca sobrescreva um par já existente sem confirmação explícita do usuário.

## Passo 1 — Extrair o conteúdo do PDF

O `Read` tool nativo precisa do `pdftoppm` (poppler), que normalmente **não está instalado** neste ambiente Windows. Use Python + `pypdf` em vez disso:

```bash
pip install pypdf --quiet   # idempotente, pula se já instalado
python -c "
from pypdf import PdfReader
reader = PdfReader(r'CAMINHO/DO/ARQUIVO.pdf')
with open('SAIDA.txt', 'w', encoding='utf-8') as f:
    for i, page in enumerate(reader.pages):
        f.write(f'\n\n===== PAGE {i+1} =====\n\n')
        f.write(page.extract_text() or '')
print(len(reader.pages), 'páginas')
"
```

Salve o `.txt` extraído no diretório de scratchpad da sessão, não no repositório.

Depois de extrair:
1. Identifique **autor/professor** do módulo — geralmente na capa (`Autoria: Nome Sobrenome`) ou nas primeiras páginas.
2. Mapeie a estrutura `Módulo NN — Título` / `Capítulo NN — Título` com grep (`^Módulo [0-9]|^Capítulo [0-9]`) para ter o índice completo antes de escrever.

## Passo 2 — Buscar material prático complementar

Cada PDF de módulo geralmente tem uma pasta de projetos correspondente no repo (ex.: `modulo04-agentes-autonomos/`, `modulo06-aiops-engenharia-agentica/`). Dentro dela, os `README.md` de cada aula/projeto contêm exemplos de código, prompts e outputs reais que **enriquecem** as seções de implementação — use-os como o "Relatório de projetos" citado no prompt padrão do Resumo. Se não existir pasta equivalente, prossiga só com o PDF.

## Passo 3 — Gerar o rascunho do Resumo (Markdown, scratchpad)

Segue **exatamente** a estrutura do `PROMPT_PADRAO_RESUMO.md` (raiz do repo — leia o arquivo, ele é o contrato formal). Resumo em pontos-chave:

### Cabeçalho
```markdown
# {Título do Módulo, igual ao nome do PDF sem extensão}
**Pós-Graduação em Engenharia de Software com IA Aplicada — UNIPDS**
{Professor} · Rafael Mattos Moreira — Resumo de estudo técnico · 2026

---
```

### Seções obrigatórias, nesta ordem, todas `##`
1. `1. Visão Geral do Módulo` — 1 parágrafo: objetivo, contexto, o que se aprende.
2. `2. Conceitos Fundamentais` — subseções `###` por conceito, com tabelas de comparação/definição e callouts de destaque.
3. `3. Detalhes de Implementação` — subseções `###` (ex.: `3.1`, `3.2`...) por projeto prático: o que faz, fluxo de dados, trechos de código com comentários em português, arquitetura (se aplicável), dependências.
4. `4. Analogias Java/Spring Boot e Go` — tabela `Conceito | Equivalente Java/Spring Boot | Equivalente Go` (três colunas, sempre as duas linguagens lado a lado, nunca só uma) + pelo menos 1 bloco de analogia arquitetural completa cobrindo ambas. Mínimo 3 analogias no módulo inteiro. O aluno trabalha com backend em Java e em Go — nunca omita a coluna Go. Quando Go não tiver equivalente direto de um recurso Java (ex.: AOP, anotações, DI automática de container), diga isso explicitamente e explique como o mesmo resultado é alcançado de forma idiomática em Go (composição explícita, middleware, interfaces) em vez de forçar uma tradução artificial.
5. `5. Escalabilidade e Produção` — como aplicar em produção real, limitações do didático, ferramentas recomendadas.
6. `6. Conexões entre Módulos` — tabela relacionando com módulos anteriores/posteriores + `### Glossário do Módulo` no final (tabela `Termo | Definição`, todo termo técnico novo do módulo).

### Formatação
- Tabelas Markdown com header em `**bold**` quando o header for um label de callout (veja próximo ponto), senão cabeçalho normal de tabela.
- Callouts como blockquote `> {emoji} **Rótulo:** texto`, usando os emojis do prompt padrão:
  - 🎯 mentalidade/dica de estudo · 💡 analogia/insight · ⚡ performance/otimização · 🔑 regra fundamental · 🔍 exemplo detalhado · 🏗️ arquitetura/fluxo · ✅ resultado/conclusão · 🔌 integração/ferramenta
- Português técnico, direto, sem inventar informação que não esteja no PDF ou no material de projeto.
- Profundidade real: explique o "porquê", não apenas liste "o quê".

## Passo 4 — Gerar o rascunho da Explicação (Markdown, scratchpad)

Documento **mais longo e narrativo** que o Resumo (tipicamente 1.5–2x mais linhas), que caminha capítulo a capítulo do PDF, na ordem em que o professor ensina — não reagrupa por tema como o Resumo faz.

### Cabeçalho
```markdown
# {Título do Módulo} — Explicação Detalhada
**Pós-Graduação em Engenharia de Software com IA Aplicada — UNIPDS**
{Professor} · Rafael Mattos Moreira · 2026

---
```
(sem "Resumo de estudo técnico" no byline — só a Explicação omite esse sufixo.)

### Estrutura hierárquica
```markdown
## Módulo N: {título do módulo tal como no PDF}

### Capítulo Y: {título do capítulo tal como no PDF}

#### {Subtópico dentro do capítulo}
```
- Um `## Módulo N` por módulo do PDF, um `### Capítulo Y` por capítulo, e quantos `####` forem necessários para separar subtemas dentro do capítulo — normalmente 3 a 8 por capítulo, dependendo do PDF ter 1 ou várias "PT-1/PT-2/PT-3" (partes) por capítulo original (agrupe as partes de um mesmo capítulo em um único `### Capítulo Y`).
- Prosa explicativa em profundidade — parágrafos completos, não bullet-only. Pode citar exemplos "ruim vs bom" com ❌/✅ quando o PDF contrastar abordagens.
- Blocos de código (` ``` `) com o que o PDF/projeto mostrar: YAML de config, JSON de exemplo, trechos de implementação, comandos de terminal.
- Os mesmos callouts do Resumo (🔑 💡 ⚡ 🎯 ✅) usados com moderação, só nos pontos de maior insight.
- Termine com uma seção de fechamento (`#### Visão Geral do Sistema Completo` ou similar) com um diagrama em bloco de texto (caixa ASCII) resumindo a arquitetura completa do módulo, mais um bloco final de analogia Java **e Go** condensada (lado a lado, mesmo formato do Passo 3) e um callout `✅` de conclusão.

## Passo 5 — Verificação antes de entregar

Releia o checklist do `PROMPT_PADRAO_RESUMO.md` (seção "Checklist de qualidade") para o Resumo. Para a Explicação, confirme adicionalmente:
- Todo `Módulo`/`Capítulo` do PDF apareceu como heading — nenhum capítulo pulado.
- Nenhuma seção é só bullet list solto sem prosa — o valor da Explicação é o texto corrido.
- Nomes de arquivo batem exatamente com o nome do PDF (sem a extensão `.pdf`), incluindo acentuação e caracteres especiais (`&`, `—` etc.).

## Passo 6 — Converter para .docx e salvar

1. Escreva o Resumo e a Explicação como Markdown puro (Passos 3 e 4) em arquivos temporários no **scratchpad da sessão** — nunca na raiz do repo.
2. Converta cada um para `.docx` com o conversor da skill:
   ```bash
   python ".claude/skills/resumo-explicacao-modulo/scripts/md_to_docx.py" "ENTRADA.md" "{Nome do Módulo} - Resumo.docx"
   python ".claude/skills/resumo-explicacao-modulo/scripts/md_to_docx.py" "ENTRADA.md" "{Nome do Módulo} - Explicação.docx"
   ```
   O script (`python-docx`) já replica o estilo visual estabelecido: Arial, headings coloridos (H1/H2/H3/H4), tabelas com grid e header sombreado, callouts em caixa (borda esquerda azul + fundo cinza-claro) e blocos de código em Courier New. Ele **não** reestrutura conteúdo — só renderiza o Markdown já escrito nos Passos 3/4.
3. Depois de gerar, abra o `.docx` com `python-docx` e confira: contagem de tabelas bate com o Markdown fonte (`grep -c "^|" ENTRADA.md`), nenhum heading ficou vazio, nenhum artefato de markdown (`**`, `` ` `` solto) sobrou fora de blocos de código.
4. Salve os dois `.docx` finais na **raiz do repositório**, ao lado do PDF de origem — mesmo local dos pares já existentes (`Criação de Agentes Autônomos - Resumo.docx`, `Model Context Protocol (MCP) - Resumo.docx`).
5. Copie os mesmos dois `.docx` também para `C:\Users\ACER\Documents\Pos em Engenharia de AI Aplicada` (pasta espelhada onde o aluno acessa os documentos fora do repo).
6. Apague os `.md` intermediários do scratchpad. Nunca deixe um `.md` de resumo/explicação solto na raiz do repo — se sobrar um, é sinal de que a conversão para `.docx` não foi concluída.
