# Model Context Protocol (MCP)
**Pós-Graduação em Engenharia de Software com IA Aplicada — UNIPDS**
Erick Wendel · Rafael Mattos Moreira — Resumo de estudo técnico · 2026

---

## 1. Visão Geral do Módulo

O módulo de MCP apresenta o Model Context Protocol como a camada de integração entre modelos de linguagem e sistemas reais (legados ou modernos). O aluno aprende a criar servidores MCP do zero, expondo **tools**, **resources** e **prompts** para que LLMs possam interagir com APIs, bancos de dados e sistemas internos de forma segura, versionada e governada. O módulo cobre desde a teoria de abstração de domínio até a publicação em NPM público/privado e integração com LangChain.js.

---

## 2. Conceitos Fundamentais

### O Problema de Integração

| Modelo | Descrição | Limitação |
|---|---|---|
| **Plugins ChatGPT** (mar/2023) | Integrações acopladas à interface | Difícil de escalar e versionar |
| **Function Calling** (jun/2023) | Funções descritas ao modelo | Descoberta manual, sem contexto amplo |
| **MCP** | Protocolo cliente-servidor com tools, resources e prompts | Exige engenharia bem-feita |

> 🔑 **Regra fundamental:** MCP não substitui APIs nem elimina sistemas legados. Ele cria uma **camada de adaptação** que permite ao modelo compreender contexto, explorar capacidades e tomar decisões informadas antes de executar ações.

### Os Três Pilares do MCP

| Pilar | Função | Analogia Java |
|---|---|---|
| **Tools** | Ações executáveis (o que o servidor FAZ) | `@Service` methods com `@Transactional` |
| **Resources** | Documentação/contexto embutido (o que o servidor É) | Swagger/OpenAPI Specification |
| **Prompts** | Templates de uso guiado (como USAR o servidor) | `@ExceptionHandler` com mensagens claras |

> 💡 **Analogia:** Tools ↔ endpoints REST. Resources ↔ Swagger doc. Prompts ↔ collection Postman pré-configurada.

### Transporte e Protocolos

| Transporte | Uso ideal | Exemplo |
|---|---|---|
| **STDIO** | Execução local, distribuição via NPM | Claude Code, VS Code extensions |
| **HTTP** | Serviço centralizado, múltiplos clientes | Gateway corporativo |
| **SSE** | Tempo real, eventos contínuos | Monitoramento, notificações |

> ⚡ **Performance:** STDIO é o padrão pois elimina overhead de rede — o servidor roda como processo local e se comunica via stdin/stdout.

### MCP vs REST

| Aspecto | REST tradicional | MCP |
|---|---|---|
| Descoberta | Manual (Swagger) | Automática (protocol negotiation) |
| Contexto | Ausente | Resources embutidos |
| Orientação | Endpoints técnicos | Ações de domínio |
| Custo tokens | Alto (enviar spec inteira) | Baixo (sob demanda) |
| Exemplo de ação | `GET /customers?name=X` | `searchCustomer({ name: "X" })` |

> 🔑 **Regra fundamental:** Não mapeie endpoints diretamente para tools. Uma tool MCP deve representar uma **intenção de negócio**, não uma rota HTTP.

---

## 3. Detalhes de Implementação

### 3.1 Projeto: Servidor MCP do Zero (CypherSuite)

**O que faz:** Criptografa e descriptografa mensagens via tools MCP com testes TDD, resources descritivos e prompts guiados.

**Fluxo de dados:**
```
Cliente MCP (editor/LangChain)
  → Tool: encrypt({ message, key })
     → CipherService.encrypt(message, key)
     → retorna { ciphertext: string }
  → Tool: decrypt({ ciphertext, key })
     → CipherService.decrypt(ciphertext, key)
     → retorna { plaintext: string }
  → Resource: info://cipher
     → Documentação do algoritmo, requisitos da chave
  → Prompt: encrypt-message({ message, key })
     → Template de instrução para o modelo
```

**Arquivos principais:**
- `server.ts` — instancia `McpServer`, define transporte STDIO
- `tools/` — registra `encrypt` e `decrypt` com schemas Zod
- `resources/` — expõe documentação técnica via URI
- `prompts/` — templates parametrizados
- `tests/` — cliente MCP em processo separado para testes de integração real

> 🔑 **TDD em MCP:** O teste cria um cliente real que se conecta ao servidor em processo separado — não é mock. Se a tool não existe, o teste falha com `tool not found`.

**Código-chave — registro de tool com schema:**
```typescript
server.tool(
  "encrypt",
  "Encrypts a message using the provided key (AES-256)",
  {
    message: z.string().describe("Plaintext message to encrypt"),
    key: z.string().min(32).describe("32-char encryption key"),
  },
  async ({ message, key }) => {
    const result = await cipherService.encrypt(message, key);
    return {
      content: [
        { type: "text", text: `Encrypted: ${result.ciphertext}` },
        { type: "resource", resource: { mimeType: "application/json", text: JSON.stringify(result) } }
      ]
    };
  }
);
```

> 💡 **Analogia Java:** `server.tool(name, description, schema, handler)` ↔ `@PostMapping` com `@RequestBody` validado por `@Valid` + Bean Validation.

---

### 3.2 Projeto: Transformando Empresa em Servidor MCP (Customers API)

**O que faz:** Expõe uma API CRUD legada como servidor MCP, adicionando busca por múltiplos critérios, autenticação JWT+RBAC, service tokens e rate limiting.

**Fluxo completo:**
```
LLM / Editor (VSCode)
  → Service Token (env var)
  → MCP Server
     → Tool: listCustomers() → HTTP Client → GET /customers
     → Tool: createCustomer({name, phone}) → HTTP Client → POST /customers
     → Tool: searchCustomer({id?, name?, phone?}) → Service (filtro em memória)
     → Tool: updateCustomer({id, name, phone}) → HTTP Client → PUT /customers/:id
     → Tool: deleteCustomer({id}) → HTTP Client → DELETE /customers/:id
     → Resource: info://api → Documentação da API legada
  → Fastify API (legada)
     → onRequest hook: verifica JWT ou Service Token
     → preHandler: RBAC (member=leitura, admin=escrita)
     → Rate limiter: 90 req/min por token
```

**Camadas da arquitetura MCP:**
```
contratos (toolbox.md, skills.md)
     ↓
MCP Server (server.ts)
     ↓
Service Layer (customerService.ts)   ← regras de negócio
     ↓
HTTP Client (apiClient.ts)           ← infraestrutura
     ↓
API Legada (Fastify + PostgreSQL)
```

> 💡 **Analogia Spring Boot:** HTTP Client ↔ `@FeignClient` ou `RestTemplate`. Service ↔ `@Service`. MCP Server ↔ `@RestController` mas voltado para LLMs, não humanos.

**Código-chave — autenticação dupla (JWT + Service Token):**
```typescript
fastify.addHook("onRequest", async (req, reply) => {
  const path = req.routerPath;
  if (PUBLIC_ROUTES.includes(path)) return; // login, health, service-token

  const token = req.headers.authorization?.replace("Bearer ", "");
  
  // Service Token
  const serviceAuth = serviceTokens.get(token);
  if (serviceAuth) {
    req.user = serviceAuth;
    return;
  }
  
  // JWT
  try {
    req.user = await fastify.jwt.verify(token);
  } catch {
    reply.status(401).send({ error: "Unauthorized" });
  }
});
```

---

### 3.3 Projeto: Múltiplos MCPs com LangChain.js (Análise de Vendas)

**O que faz:** Agente autônomo que recebe dados CSV/JSON, identifica intenção, converte formato, persiste no MongoDB e gera relatório — usando 3 MCPs em paralelo.

**Pipeline completo:**
```
Usuário: { question: "rank produtos mais vendidos", data: "..." }
  ↓ Agente de Intenção (structured output)
  → { intent, fileType, content, fileName }
  ↓ (se fileType === "csv")
  → Tool customizada: csvToJson(csvContent)    ← biblioteca determinística
  ↓ Agente Executor
  → MongoDB MCP: dropCollection, insertMany, aggregate
  → File System MCP: writeFile(report.txt)
  → retorna relatório final
```

**Ferramentas disponíveis:**
| Ferramenta | Tipo | Responsabilidade |
|---|---|---|
| `csvToJson` | Custom LangChain Tool | Conversão determinística CSV→JSON |
| MongoDB MCP | Servidor MCP externo | CRUD + Aggregation |
| File System MCP | Servidor MCP externo | Leitura/escrita de arquivos |

> 🎯 **Mentalidade:** Tarefas determinísticas (converter CSV) → tool customizada. Operações externas (banco, arquivo) → MCP padronizado. Não deixe a LLM executar cálculos que código faz melhor.

---

### 3.4 Projeto: Google Trends como Tool (Services como Tools)

**O que faz:** Demonstra o padrão de expor services existentes como tools LangChain, enriquecendo respostas com dados reais de tendências.

**Fluxo:**
```
Usuário: "sugestões de títulos sobre Node.js performance"
  → Nó pesquisa (LLM com Google Trends tool)
     → extrai keywords: ["Node.js performance", "V8 optimization"]
     → chama 1x (apenas!) GoogleTrendsService.getTrends(keywords)
     → retorna tópicos em alta, queries relacionadas
  → Nó resposta (LLM sem tools)
     → usa dados de tendência para gerar títulos relevantes
```

> ⚡ **Performance:** Forçar 1 chamada via prompt evita exploração excessiva em APIs pagas — padrão crítico em produção.

---

### 3.5 Projeto: Publicação MCP (NPM + Verdaccio)

**O que faz:** Empacota e distribui o servidor MCP como pacote NPM, primeiro em registry privado (Verdaccio) para testes, depois no NPM público.

**Configuração binário:**
```json
// package.json
{
  "bin": { "customers-mcp": "./src/server.ts" },
  "scripts": { "publish:private": "npm publish --registry http://localhost:4873" }
}
```

**Uso após publicação:**
```json
// .vscode/mcp.json
{
  "servers": {
    "customers": {
      "command": "npx",
      "args": ["-y", "customers-mcp-server"],
      "env": { "API_SERVICE_TOKEN": "${env:CUSTOMERS_TOKEN}" }
    }
  }
}
```

---

### 3.6 Projeto: MCP + LangChain.js (Agente Completo)

**O que faz:** Integra o Customers MCP publicado em uma aplicação LangChain.js como dependência real — o agente gerencia clientes via linguagem natural.

**Fluxo:**
```
"criar 3 clientes de teste e salvar em arquivo"
  → LangChain Agent
     → Customers MCP (createCustomer × 3)
     → File System MCP (writeFile customers.json)
     → listCustomers() → retorno consolidado
```

> ✅ **Resultado:** O MCP deixa de ser protótipo e se torna dependência versionada reutilizável em qualquer aplicação.

---

## 4. Analogias Java/Spring Boot

| Conceito MCP | Equivalente Java/Spring Boot | Detalhe |
|---|---|---|
| `server.tool()` | `@PostMapping` + `@RequestBody @Valid` | Schema Zod ↔ Bean Validation |
| `server.resource()` | `@GetMapping` + Swagger `@ApiResponse` | Documentação embutida |
| `server.prompt()` | Postman Collection compartilhada | Template de uso |
| STDIO transport | Process fork (ProcessBuilder) | Processo filho com I/O streams |
| HTTP transport | `@RestController` exposto | Endpoint HTTP normal |
| Service Token | `@PreAuthorize` + JWT custom | Token de longa duração |
| Rate Limiting | `@RateLimiter` (Resilience4j) | `429 Too Many Requests` |
| RBAC (member/admin) | `@Secured({"ROLE_ADMIN"})` | Spring Security roles |
| `onRequest` hook | `HandlerInterceptor.preHandle()` | Executa antes do endpoint |
| Zod schema | `javax.validation` annotations | Validação de entrada |

> 💡 **Analogia-chave:** Um servidor MCP completo (tools + resources + prompts) ↔ um `@RestController` com Swagger bem documentado, autenticação configurada e collection Postman disponível. A diferença é que o consumidor é uma LLM, não um humano.

---

## 5. Escalabilidade e Produção

### Limitações do Approach Didático

| Aspecto | Didático | Produção |
|---|---|---|
| Service Tokens | Armazenados em memória | Redis/banco persistente |
| Rate limiting | Por token apenas | Token + IP |
| Autenticação | Usuários em memória | Identity Provider (Keycloak, Auth0) |
| Secrets | `.env` local | Vault, AWS Secrets Manager |
| MCP transport | STDIO local | HTTP + TLS para múltiplos clientes |

### Segurança em MCPs Públicos

```
✅ SEMPRE FAZER:
  - Validar inputs com schema (Zod/Joi)
  - Service Tokens via env vars
  - Rate limiting por token E por IP
  - Apenas operações read-only em dados críticos
  - Repositórios oficiais para MCPs de terceiros

❌ NUNCA FAZER:
  - Hardcode de secrets no código
  - Mapeamento 1:1 endpoint→tool (vaza estrutura interna)
  - Sem rate limiting em APIs pagas
  - MCPs sem autenticação em produção
```

### Padrão de Rate Limiting em MCP

```typescript
// Identificação por token (auth) ou IP (fallback)
keyGenerator: (req) => {
  const token = req.headers.authorization?.replace("Bearer ", "");
  return token || req.ip;
}
// Resposta padrão: 429 + Retry-After header
```

### Recomendações para Scale

- **Múltiplos clientes simultâneos** → HTTP transport + autoscaling
- **Dados grandes** → Não enviar no prompt; usar tools de leitura sob demanda
- **Context window** → Limite de tokens define volume máximo de dados processáveis
- **Custo** → MCP reduz tokens vs REST (ações em vez de spec completa)

---

## 6. Conexões entre Módulos

| Módulo anterior | Como MCP se conecta |
|---|---|
| **Módulo 01** (TensorFlow, Embeddings) | MCP pode expor modelos como tools (`predictCategory`, `searchSimilar`) |
| **Módulo 02** (APIs LLMs, RAG) | MCP é a camada de abstração que o RAG usa para buscar dados |
| **Módulo 03** (MCP — este) | Base para os agentes do próximo módulo |
| **Módulo 04** (Agentes) | Agentes autônomos consomem MCPs como ferramentas |

> 🏗️ **Arquitetura macro:**
> ```
> LLM / Agente Autônomo
>      ↓ usa
> MCP Server (tools, resources, prompts)
>      ↓ abstrai
> API Legada / Banco / Serviço Interno
> ```

### Glossário do Módulo

| Termo | Definição |
|---|---|
| **MCP** | Model Context Protocol — padrão de comunicação LLM↔ferramentas |
| **Tool** | Função executável registrada no servidor MCP |
| **Resource** | Documento/contexto embutido no servidor (não executa ação) |
| **Prompt** | Template de instrução parametrizado |
| **STDIO** | Standard Input/Output — transporte padrão para MCPs locais |
| **Service Token** | API key de longa duração para integrações programáticas |
| **RBAC** | Role-Based Access Control — autorização por papel |
| **Rate Limiting** | Controle de volume de requisições por período |
| **Zod** | Biblioteca TypeScript de validação e tipagem de schemas |
| **Tool discovery** | Capacidade do cliente MCP explorar automaticamente as tools disponíveis |
