# Model Context Protocol (MCP) — Explicação Detalhada
**Pós-Graduação em Engenharia de Software com IA Aplicada — UNIPDS**
Erick Wendel · Rafael Mattos Moreira · 2026

---

## Módulo 1: Visão Geral de MCP

### Capítulo 1: Diferença entre MCPs e o Modelo Clássico de Tools/Plugins

#### O Problema de Integração

Antes do MCP, cada integração com LLMs exigia:
- Documentação extensa para cada API
- Múltiplos endpoints com autenticação específica
- Manutenção constante à medida que as integrações mudam

Esse problema se repetiu no contexto de aplicações com modelos de linguagem.

**Evolução histórica:**

| Data | Tecnologia | O que fazia |
|---|---|---|
| Mar/2023 | Plugins ChatGPT | Integrações acopladas à interface, permitia buscar dados externos |
| Jun/2023 | Function Calling | Funções descritas ao modelo com nome, descrição e schema de entrada |
| Hoje | MCP | Protocolo cliente-servidor com tools, resources e prompts + descoberta automática |

#### Limitações do Function Calling

O Function Calling resolveu o problema de integração nas **aplicações**, mas tinha limitações:

1. **Entendimento superficial**: O modelo sabe que a função existe, mas não tem visão ampla do contexto
2. **Trabalho repetitivo**: Descrever manualmente todas as funções no código
3. **Sem mecanismos padronizados de descoberta**: O modelo não consegue "investigar" a integração — apenas consome o que foi definido

#### O Que o MCP Propõe de Diferente

O MCP eleva o nível de abstração:

- **Protocolo de comunicação** (não apenas lista de funções)
- **Descoberta automática** — o cliente explora recursos disponíveis sem configuração manual
- **Contexto rico** — resources descrevem o serviço; prompts guiam o uso
- **Ações de domínio** em vez de endpoints técnicos

> 🔑 Em vez de `GET /customers?id=123`, o MCP expõe `searchCustomer({ id: "123" })` — internamente pode chamar vários endpoints, autenticar, agregar e devolver uma resposta consolidada.

#### MCP Substitui Tools?

**Não.** O MCP incorpora tools dentro de um protocolo mais amplo. Muitos frameworks continuam usando o termo "tools" mesmo quando operam com MCP. A diferença é que por trás existe um protocolo mais rico organizando essas capacidades.

#### Comparação com APIs Tradicionais

| Aspecto | REST/Swagger | MCP |
|---|---|---|
| Documentação | Arquivo externo (Swagger) | Embutida no servidor (resources) |
| Custo de tokens | Alto — enviar spec inteira | Baixo — consumo sob demanda |
| Descoberta | Manual | Automática via protocolo |
| Orientação | Técnica (endpoints) | Negócio (ações de domínio) |

> ⚡ MCP resolve o problema de custo ao trabalhar com ações em vez de endpoints — o modelo consome apenas o necessário, solicitando mais sob demanda.

#### Engenharia Continua Sendo Essencial

Um servidor MCP mal projetado pode ser lento, inseguro e difícil de manter, assim como qualquer outro sistema. O protocolo facilita, mas **não substitui decisões arquiteturais bem feitas**.

---

## Módulo 2: Conectando LLMs a APIs, Bancos e Serviços

### Capítulo 1: Template Inicial e Arquitetura

#### Objetivo da Aplicação

O projeto simula análise de dados a partir de relatório de vendas (CSV/JSON). O diferencial: **o usuário não especifica formato, estrutura ou etapas intermediárias** — o modelo orquestra tudo.

**Mudança de paradigma:**
- Antes: controle rígido de fluxo via código imperativo
- Depois: modelo recebe capacidades + instrução; decide quando e como usá-las

#### Pipeline de Execução (orientado pelo modelo)

```
1. Conversão de dados (se CSV → JSON via tool customizada)
2. Persistência opcional em arquivo (File System MCP)
3. Inserção no MongoDB (MongoDB MCP)
4. Processamento e análise (queries de agregação)
5. Geração de relatório (escrita em arquivo)
```

> 🎯 **Insight-chave:** O pipeline não é rigidamente programado. É **descrito em alto nível** e o modelo decide como executá-lo.

#### Ferramentas Disponíveis

| Ferramenta | Tipo | Quando usar |
|---|---|---|
| File System | MCP | Ler/escrever arquivos em disco |
| CSV→JSON Converter | Custom Tool LangChain | Conversão determinística de formato |
| MongoDB | MCP | Armazenar, consultar e agregar dados |

#### Tratamento de Falhas e Retentativas

Se o modelo executar uma operação incorreta (query inválida), ele pode:
1. Consultar novamente o servidor MCP
2. Buscar exemplos de uso nos resources
3. Ajustar a execução
4. Tentar novamente

Isso é **resiliência autônoma** — sem intervenção manual.

#### Considerações sobre Contexto e Desempenho

> ⚡ O limite de contexto da LLM determina o volume máximo de dados processáveis em uma única execução. Para volumes maiores: fazer upload de arquivos e usar tools de leitura sob demanda.

---

### Capítulo 2: Parsing Inteligente — Organizando a Entrada do Cliente

#### Por que Parsing Inteligente é Crítico

Qualquer erro na interpretação inicial compromete todo o restante do pipeline. O nó de intenção transforma entrada não estruturada em objeto estruturado:

```typescript
// Entrada do usuário
"rank dos produtos mais vendidos do CSV abaixo: produto,qtd\nA,10\nB,5"

// Saída estruturada (geração estruturada com schema)
{
  intent: "rank produtos mais vendidos",
  fileType: "csv",
  content: "produto,qtd\nA,10\nB,5",
  fileName: "sales-data"
}
```

#### Uso de Geração Estruturada

Em vez de texto livre, o modelo retorna JSON que segue um schema estrito. Isso:
- Reduz ambiguidades
- Permite validação automática
- Garante que campos obrigatórios existam

#### Tratamento de Inconsistências

Se o modelo não consegue gerar um nome de arquivo, o sistema usa fallback baseado no fileType. Essa resiliência garante que o fluxo continue mesmo com pequenas inconsistências.

> 🔑 Se a intenção não for identificada → **interrompe o fluxo imediatamente**. Não faz sentido continuar o processamento sem informações básicas.

---

### Capítulo 3: Orquestração Autônoma com LangChain.js e MCP no MongoDB

#### Por que Não Deixar a LLM Executar Tudo

Embora o modelo consiga converter CSV e fazer agregações por conta própria:
- Não foi projetado para cálculos complexos com **precisão garantida**
- O processamento dentro do modelo **consome tokens** desnecessariamente

> 🔑 A estratégia correta é **delegar para ferramentas externas** tudo que é determinístico.

#### Configuração do MongoDB como Tool

```typescript
// configuração do MCP
const mongoMcp = {
  name: "mongodb",
  transport: "stdio",
  command: "npx",
  args: ["@mongodb/mcp-server", "--database", "sales_db"],
  permissions: { read: true, write: true }
};
```

#### Execução Autônoma na Prática

Com a tool integrada, o modelo passa a:
1. Limpar o banco automaticamente (evitar dados antigos)
2. Inserir os dados recebidos
3. Executar consultas de agregação
4. Gerar resultados

**Tudo sem código imperativo definindo cada passo.**

#### Observabilidade do Processo

Através dos logs é possível acompanhar:
- Quando o modelo decide chamar uma tool
- Qual ferramenta foi utilizada e com quais parâmetros
- Qual foi o resultado da execução
- Tentativas e retentativas

---

### Capítulo 4: Construindo uma Tool Customizada no LangChain

#### Por que Criar Tool Customizada em Vez de Usar a LLM

| Problema com LLM | Solução com Tool |
|---|---|
| Inconsistências na conversão | Código determinístico sempre correto |
| Variação de resultado com contexto | Mesmo input → mesmo output |
| Gasta tokens desnecessariamente | Zero tokens para tarefa técnica |
| Tarefa probabilística | Tarefa determinística |

#### Estrutura de uma Tool LangChain

```typescript
import { tool } from "@langchain/core/tools";
import { z } from "zod";
import csvToJsonLib from "csv-to-json";

const csvToJsonTool = tool(
  async ({ csvContent }) => {
    const result = await csvToJsonLib.parse(csvContent);
    console.log(`Converted ${result.length} rows`);
    return JSON.stringify(result); // LLM prefere string
  },
  {
    name: "csv_to_json",
    description: "Converts CSV text to JSON array. Use this instead of manual conversion.",
    schema: z.object({
      csvContent: z.string().describe("Raw CSV content as string")
    })
  }
);
```

> 🔑 **A descrição importa**: "Use this instead of manual conversion" instrui explicitamente o modelo a não tentar converter sozinho.

#### Observando a Chamada em Tempo Real

Ver o modelo chamar a tool no debugger remove a "mágica":
1. Modelo analisa o problema
2. Identifica que precisa converter CSV
3. Chama a ferramenta com os argumentos corretos
4. Aguarda o retorno para continuar

#### Diferença entre Tool Customizada e Servidor MCP

| Tool Customizada | Servidor MCP |
|---|---|
| Função local na aplicação | Processo separado com protocolo |
| Sem descoberta automática | Discovery automático |
| Sem resources/prompts | Contexto rico embutido |
| Simples, para tarefas pontuais | Para integrações complexas e reutilizáveis |

---

### Capítulo 5: File System MCP

#### Por Que o Acesso a Arquivos Transforma o Sistema

Antes: sistema responde perguntas e manipula dados em memória.
Depois: sistema **gera artefatos persistentes**, organizados em diretórios, auditáveis.

#### Controle de Escopo

> ⚡ **Crítico:** Restringir o diretório acessível ao modelo.

```typescript
// Acesso irrestrito ao projeto = problema
{ allowedPaths: ["/"] }

// Correto: apenas pasta de relatórios
{ allowedPaths: ["/app/reports"] }
```

Sem escopo restrito, o modelo pode tentar explorar arquivos desnecessários, aumentando consumo de tokens e perdendo foco.

#### Comparação entre Tools e MCP (Resumo)

| Aspecto | Tools Customizadas | Servidor MCP |
|---|---|---|
| Natureza | Funções locais | Protocolo + processo separado |
| Descoberta | Manual | Automática |
| Contexto | Ausente | Resources + prompts |
| Uso ideal | Tarefas pontuais | Integrações complexas/reutilizáveis |

---

### Capítulo 6: Google Trends API como Tool (Services como Tools)

#### O Padrão de Expor Services como Tools

Quando existe uma service com regras de negócio específicas da empresa, em vez de criar um MCP completo:

1. Implementar a service normalmente (lógica de negócio encapsulada)
2. Expor a service como uma tool LangChain
3. O modelo decide **quando** usar — a lógica permanece encapsulada

**Benefícios:**
- Autonomia da LLM para decidir
- Controle da aplicação sobre a lógica
- Reaproveitamento de código existente
- Menor acoplamento entre fluxo e implementação

#### Controle de Custo em APIs Pagas

> ⚡ APIs com limite de uso: forçar **1 chamada apenas** via prompt. Sem essa instrução, o modelo pode fazer chamadas exploratórias repetidas.

```
System Prompt: "Execute only ONE call to the trends tool. 
Do not call it multiple times regardless of the results."
```

---

## Módulo 3: MCPs vs Tools para Ambiente de Desenvolvimento

### Capítulo 1: Agents e Instructions

#### Vibe Coding e a Necessidade de Contexto

No cenário atual, cada vez mais delegamos tarefas para IA. A qualidade do resultado depende **de como instruímos o modelo**, não de quanto digitamos.

Dois elementos fundamentais:
1. **Arquivos de instrução** — descrevem o projeto para a IA
2. **Agents especializados** — prompts com responsabilidades bem definidas

#### Arquivos de Instrução (CLAUDE.md, .cursorrules, etc.)

```markdown
# Contexto do Projeto

## Arquitetura
- TypeScript + Node.js
- Clean Architecture (Use Cases, Repositories, Entities)
- Testes com Vitest

## Regras
- Imutabilidade em todas as funções
- Sem `any` implícito
- Testes obrigatórios para todos os use cases
```

> 🔑 Esses arquivos funcionam como guia permanente. A IA não precisa reaprender o contexto a cada prompt.

#### O Problema do Tamanho e Consumo de Tokens

Arquivos de instrução muito grandes:
- Maior consumo de tokens por requisição
- Maior risco de a IA "se perder" no contexto
- Respostas menos precisas

Solução: usar **agents especializados** com prompts menores e focados.

#### O Padrão llms.txt

Similar ao `robots.txt`, mas para LLMs:
```
# llms.txt
> Descrição concisa do sistema para agentes de IA

- [API Docs](./docs/api.md) — documentação da API REST
- [Architecture](./docs/arch.md) — decisões arquiteturais
```

Reduz consumo de recursos ao fornecer estrutura navegável em vez de HTML.

#### Agents na Prática

| Agent | Responsabilidade | Tools |
|---|---|---|
| Dev Agent | Gerar/editar código TypeScript | File read/write, terminal |
| Test Agent | Criar testes unitários/integração | File read/write |
| Review Agent | Revisar código contra padrões | File read |
| Deploy Agent | CI/CD, criação de PR | Git, GitHub API |

> 🎯 **Regra:** Critérios de "done" do agent: sem erros de compilação + testes passando + alterações validadas.

---

### Capítulo 2: Skills

#### O Problema dos Prompts Grandes

Quanto maior o prompt, maior a chance da IA se perder:
- Limite de tokens
- O modelo não consegue priorizar todas as informações
- Partes importantes podem ser ignoradas
- Resposta perde consistência

#### O Que São Skills

Skills são unidades menores de conhecimento especializado:

```markdown
# Skill: PostgreSQL Performance Queries

## Quando usar
Ao escrever queries para tabelas com > 1M de registros

## Padrões
- Sempre usar índices parciais para colunas com alta cardinalidade
- Evitar SELECT * em produção
- Usar EXPLAIN ANALYZE antes de fazer deploy

## Exemplo
✅ SELECT id, name FROM users WHERE active = true AND created_at > NOW() - INTERVAL '30 days'
❌ SELECT * FROM users WHERE active = true
```

#### Comparação Skills vs MCP

| Aspecto | Skills | MCP |
|---|---|---|
| O que organiza | **Conhecimento** (como fazer) | **Ferramentas** (o que executar) |
| Filosofia | Instrução | Execução |
| Consumo | Carregado no prompt | Chamado sob demanda |
| Descoberta | Auto (sistema identifica relevância) | Manual ou auto via protocolo |

> 💡 **Analogia:** Skills ↔ manual técnico interno. MCP ↔ sistema que executa o que o manual descreve.

#### Combinação Skills + Agents + MCP

```
Agent (papel) 
  usa Skills (conhecimento especializado)
  acessa MCPs (ferramentas externas)
  
Exemplo:
  Dev Agent 
    → Skill: Clean Architecture patterns
    → Skill: TypeScript best practices
    → MCP: GitHub (criar PR)
    → MCP: Jira (atualizar ticket)
```

---

## Módulo 4: CypherSuite — Desenvolvendo MCPs do Zero

### Capítulo 1: Criando um MCP do Zero com TDD

#### Por que Começar pelos Testes

```typescript
// Teste cria um CLIENTE REAL que se conecta ao servidor em processo separado
const client = await createTestClient({
  command: "npx",
  args: ["ts-node", "src/server.ts"]
});

// Teste falha porque a tool ainda não existe — isso é intencional
it("should encrypt a message", async () => {
  const result = await client.callTool("encrypt", {
    message: "hello world",
    key: "12345678901234567890123456789012"
  });
  expect(result.content[0].text).toContain("Encrypted:");
});
```

> 🔑 Não estamos testando funções internas — estamos validando a **integração real** entre cliente e servidor, como aconteceria em produção.

#### Estrutura do Servidor

```typescript
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";

const server = new McpServer({
  name: "cypher-suite",
  version: "1.0.0"
});

// STDIO: processo local, zero infraestrutura
const transport = new StdioServerTransport();
await server.connect(transport);
```

#### Registrando uma Tool com Schema

```typescript
server.tool(
  "encrypt",
  "Encrypts a message using AES-256-CBC with the provided key",
  {
    message: z.string().min(1).describe("Plaintext to encrypt"),
    key: z.string().length(32).describe("32-character encryption key")
  },
  async ({ message, key }) => {
    try {
      const result = await cipherService.encrypt(message, key);
      return {
        content: [
          { type: "text", text: `Encrypted: ${result.ciphertext}` },
          { type: "resource", resource: { 
            mimeType: "application/json", 
            text: JSON.stringify(result) 
          }}
        ]
      };
    } catch (err) {
      return {
        isError: true,
        content: [{ type: "text", text: `Error: ${err.message}` }]
      };
    }
  }
);
```

> 🔍 **Dois formatos de retorno:** texto (para LLMs simples) + JSON estruturado (para processamento automático). Isso maximiza compatibilidade.

#### Inspecionando o Servidor MCP

```bash
# MCP Inspector — ferramenta visual para testar servidores
npx @modelcontextprotocol/inspector npx ts-node src/server.ts
```

O inspector mostra as 3 áreas do protocolo: **Tools**, **Resources**, **Prompts** — mesmo que ainda não implementadas.

---

### Capítulo 2: Resources, Prompts e Integração com VSCode

#### Registrando um Resource

```typescript
server.resource(
  "cipher-info",
  "info://cipher",
  "Technical documentation about the cipher implementation",
  async () => ({
    contents: [{
      uri: "info://cipher",
      mimeType: "text/plain",
      text: `
# CypherSuite MCP — Technical Reference

## Algorithm
AES-256-CBC symmetric encryption

## Key Requirements  
- Exactly 32 characters
- Store securely; never log or expose

## Output Format
{ ciphertext: string, iv: string }

## Usage Notes
Use the same key for encryption and decryption.
Different keys produce different (unrecoverable) results.
      `.trim()
    }]
  })
);
```

> 💡 Resources são como documentação embutida no servidor. A LLM pode ler antes de executar tools, tomando decisões melhores.

#### Registrando um Prompt

```typescript
server.prompt(
  "encrypt-message",
  "Template to encrypt a message using the cipher service",
  {
    message: z.string().describe("Message to encrypt"),
    key: z.string().describe("Encryption key (32 chars)")
  },
  ({ message, key }) => ({
    messages: [{
      role: "user",
      content: {
        type: "text",
        text: `Please encrypt the following message using the encrypt tool:
Message: "${message}"
Key: "${key}"
Return the encrypted result.`
      }
    }]
  })
);
```

#### Conectando ao VSCode

```json
// .vscode/mcp.json
{
  "servers": {
    "cypher-suite": {
      "command": "npx",
      "args": ["ts-node", "${workspaceFolder}/src/server.ts"]
    }
  }
}
```

**O que acontece:**
1. VSCode detecta a configuração
2. Inicia o processo MCP
3. Descobre tools, resources e prompts automaticamente
4. Permite uso interativo no editor

#### Fluxo Completo no Editor

```
Usuário seleciona Prompt "encrypt-message"
  → Editor pede: message, key
  → Monta instrução: "Please encrypt..."
  → Envia para o modelo
  → Modelo chama tool: encrypt({ message, key })
  → Servidor executa criptografia
  → Retorna resultado para o editor
```

> ✅ **Resultado:** Tools + Resources + Prompts formam uma interface completa que vai além de uma lista de funções.

---

## Módulo 5: Transformando Empresa em Servidor MCP

### Capítulo 2: Como Empresas Usam MCPs para Sistemas Legados

#### Por Que Não Mapear Endpoints Diretamente

```
❌ ANTI-PATTERN: Mapeamento 1:1 com REST
  tool: getCustomer(id)     → GET /customers/:id
  tool: listCustomers()     → GET /customers
  tool: getCustomerOrders() → GET /customers/:id/orders

✅ CORRETO: Abstração de domínio
  tool: findCustomerWithHistory(name) → 
    1. GET /customers?name=X
    2. GET /customers/:id/orders (para cada cliente)
    3. Consolida e retorna visão completa
```

> 🔑 O MCP deve **esconder a complexidade** da API original — múltiplos endpoints, autenticação, paginação — e expor apenas a **intenção de negócio**.

#### Arquitetura em Camadas

```
MCP Server (interface para LLMs)
     ↓
Service Layer (regras de negócio, agregações)
     ↓
HTTP Client (comunicação com API legada)
     ↓
API Legada (intocada)
```

> 💡 **Analogia Spring Boot:** HTTP Client ↔ `@FeignClient`. Service ↔ `@Service`. MCP Server ↔ novo "controller" para LLMs.

#### Resource como Documentação Viva

```typescript
server.resource("api-docs", "info://api", "Legacy API documentation", async () => ({
  contents: [{
    uri: "info://api",
    mimeType: "text/plain",
    text: `
Base URL: http://localhost:3001
Endpoints: GET /customers, POST /customers, PUT /customers/:id, DELETE /customers/:id
Auth: Bearer token required (except health check)
    `
  }]
}));
```

---

### Capítulos 3 e 4: CRUD Completo de Clientes

#### Busca com Múltiplos Critérios (Abstração além da API)

A API original não possui busca flexível. O MCP adiciona essa capacidade:

```typescript
// Service Layer — lógica que a API original não tem
async searchCustomers(query: Partial<Customer>): Promise<Customer[]> {
  if (query.id) {
    const customer = await this.client.getCustomerById(query.id);
    return customer ? [customer] : [];
  }
  
  const all = await this.client.listCustomers();
  return all.filter(customer => 
    Object.entries(query).every(([key, value]) => 
      customer[key]?.toString().toLowerCase().includes(value.toLowerCase())
    )
  );
}
```

> ✅ **Valor real:** O MCP adicionou capacidade de busca que a API original não tinha — sem modificar a API.

---

## Módulo 6: Segurança e Governança em MCPs

### Capítulo 2: JWT + RBAC em Web APIs

#### Autenticação vs Autorização

| Conceito | Pergunta | Implementação |
|---|---|---|
| **Autenticação** (JWT) | Quem é você? | Login → token assinado |
| **Autorização** (RBAC) | O que você pode fazer? | Role no token → verifica permissão |

#### Hook Global de Autenticação (Fastify)

```typescript
fastify.addHook("onRequest", async (req, reply) => {
  if (PUBLIC_ROUTES.includes(req.routerPath)) return;
  
  try {
    req.user = await fastify.jwt.verify(
      req.headers.authorization?.replace("Bearer ", "")
    );
  } catch {
    reply.status(401).send({ error: "Unauthorized" });
  }
});
```

> ⚡ **Segurança por padrão:** Tudo é privado, exceto o que foi explicitamente liberado. Reduz risco de esquecer uma rota desprotegida.

#### RBAC com preHandler

```typescript
const requireRole = (role: "admin" | "member") => async (req, reply) => {
  if (req.user.role !== role && !(role === "member" && req.user.role === "admin")) {
    reply.status(403).send({ error: "Insufficient permissions" });
  }
};

// Aplicando nas rotas
fastify.post("/customers", { preHandler: requireRole("admin") }, createHandler);
fastify.get("/customers", { preHandler: requireRole("member") }, listHandler);
```

---

### Capítulo 3: Service Tokens para MCPs

#### Por que MCPs Precisam de um Modelo Diferente

| JWT (usuários humanos) | Service Token (MCPs/integrações) |
|---|---|
| Expira (exige renovação) | Não expira automaticamente |
| Representa sessão de usuário | Representa integração/aplicação |
| Fluxo de login | Configuração única |
| Armazenamento stateless | Requer armazenamento server-side |

#### Geração de Service Token

```typescript
// Rota protegida por SUPER_SECRET
fastify.post("/auth/service-token", async (req, reply) => {
  const { username, password, adminSuperSecret } = req.body;
  
  if (adminSuperSecret !== process.env.SUPER_ADMIN_SECRET) {
    return reply.status(401).send({ error: "Invalid super secret" });
  }
  
  const user = validateUser(username, password); // reutiliza lógica de autenticação
  if (!user) return reply.status(401).send({ error: "Invalid credentials" });
  
  const token = crypto.randomUUID() + crypto.randomUUID(); // UUID longo e imprevisível
  serviceTokens.set(token, { username: user.username, role: user.role });
  
  return { token, role: user.role };
});
```

---

### Capítulo 4: Rate Limiting

#### Por que Rate Limiting é Crítico para MCPs

MCPs podem automatizar chamadas, encadear operações e executar tarefas em sequência. **Sem controle, isso pode gerar cargas muito altas.**

```typescript
await fastify.register(fastifyRateLimit, {
  max: 90,  // 90 requisições
  timeWindow: "1 minute",
  keyGenerator: (req) => {
    const token = req.headers.authorization?.replace("Bearer ", "");
    return token || req.ip;  // por token (auth) ou IP (fallback)
  }
});
```

**Resposta ao exceder o limite:**
```json
{ "statusCode": 429, "error": "Too Many Requests", "message": "Rate limit exceeded, retry in 45 seconds" }
```

> ⚡ **Cada token tem seu próprio limite.** Dois clientes diferentes não competem entre si — cada integração tem seu próprio controle de uso.

---

## Módulo 7: MCP em Produção

### Capítulo 1: Publicando em NPM e Verdaccio

#### Workflow de Publicação

```bash
# 1. Testar no registry privado (Verdaccio)
npm adduser --registry http://localhost:4873
npm version patch
npm publish --registry http://localhost:4873

# 2. Validar uso via npx
npx --registry http://localhost:4873 customers-mcp-server

# 3. Publicar no NPM público (após validação)
npm login
npm publish
```

**Convenção de versionamento semântico:**
- `patch` (1.0.1): correção de bugs
- `minor` (1.1.0): novas funcionalidades sem quebra
- `major` (2.0.0): mudanças incompatíveis

#### Tornar o MCP Executável via npx

```json
// package.json
{
  "name": "customers-mcp-server",
  "bin": {
    "customers-mcp": "./src/server.ts"
  }
}
```

```typescript
// src/server.ts — primeira linha obrigatória
#!/usr/bin/env node
```

> 💡 **Analogia:** Distribuir um servidor MCP via NPM ↔ distribuir um JAR executável via Maven Central. Qualquer projeto pode adicionar como dependência e usar via `npx`.

---

### Capítulo 2: Diferentes Transports

#### Quando Usar Cada Transport

| Transport | Ideal para | Exemplo de uso |
|---|---|---|
| **STDIO** | Ferramentas locais, CLIs, editores | Claude Code, VS Code extensions |
| **HTTP** | Serviços centralizados, múltiplos clientes | Gateway MCP corporativo |
| **HTTP Streaming** | Dados incrementais, processamento longo | Processamento de vídeo em partes |
| **SSE** | Tempo real, eventos contínuos | Monitoramento, notificações ao vivo |
| **Docker** | Ambientes padronizados, dependências complexas | Infraestrutura corporativa |

> 🔑 **Trade-off:** STDIO = simplicidade + segurança (sem exposição de rede). HTTP = escalabilidade + acesso remoto. Para a maioria dos casos, STDIO via NPM é a escolha certa.

---

## Módulo 8: MCP com LangChain.js

### Capítulo 1: Agente com Customers MCP

#### Consumindo MCP Publicado como Dependência

```typescript
// Usa o pacote publicado — não o código-fonte local
const customersMcp = {
  command: "npx",
  args: ["-y", "customers-mcp-server"],
  env: { API_SERVICE_TOKEN: process.env.CUSTOMERS_TOKEN }
};
```

#### Grafo do Agente

```typescript
const graph = new StateGraph(AgentState)
  .addNode("agent", agentNode)
  .addEdge(START, "agent")
  .addEdge("agent", END);
```

**O agente recebe múltiplos MCPs:**
- `customersMcp` — operações de clientes
- `fileSystemMcp` — persistência de resultados

#### Resultado Real da Autonomia

```
Input: "criar 3 clientes de teste e listar todos depois"

Execução autônoma:
  1. createCustomer({ name: "Test 1", phone: "11999990001" })
  2. createCustomer({ name: "Test 2", phone: "11999990002" })
  3. createCustomer({ name: "Test 3", phone: "11999990003" })
  4. listCustomers() → retorna todos incluindo os 3 novos

Output: "Foram criados 3 clientes. Total de X clientes cadastrados: [lista]"
```

#### Por que Memória de Conversa é Necessária

Sem memória: o agente não sabe que acabou de criar os clientes quando perguntado sobre IDs específicos.

Com memória de conversa (`MessagesAnnotation`): cada turno acumula histórico, permitindo referências ao contexto anterior.

> ✅ **Conclusão do módulo:** O MCP não é apenas um projeto técnico. É uma **capacidade viva da aplicação** — um bloco reutilizável, versionado, testado e seguro que qualquer agente pode consumir.
