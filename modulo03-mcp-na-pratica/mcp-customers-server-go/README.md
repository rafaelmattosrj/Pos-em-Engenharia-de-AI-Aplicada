# MCP Customers Server em Go

Porte em Go do projeto Spring Boot [`mcp-customers-server-java`](../mcp-customers-server-java) — servidor **MCP (Model Context Protocol)** que encapsula a [`legacy-customer-api`](../legacy-customer-api-go), expondo tools, um resource de documentação e prompts via transporte stdio, para uso em Claude Desktop e outros clientes MCP.

## O que o servidor expõe

**5 tools:**
- `listCustomers` — lista todos os clientes
- `searchCustomer` — busca por `id`/`name`/`phone` (filtro em memória, parâmetros opcionais)
- `createCustomer` — cria um cliente (`name`, `phone`)
- `updateCustomer` — atualiza um cliente (`id` obrigatório, `name`/`phone` opcionais)
- `deleteCustomer` — remove um cliente (`id`)

**1 resource:**
- `info://api` — documentação em Markdown da legacy customer API (endpoints, autenticação, formato dos dados)

**2 prompts:**
- `search-customer-prompt` — template para orientar o LLM a buscar clientes a partir de uma query livre
- `create-customer-prompt` — template para orientar o LLM a cadastrar um cliente com validação básica

## O que foi mantido 1:1

- Mesmas 5 tools, com as mesmas descrições e os mesmos parâmetros (obrigatórios/opcionais).
- Mesmo resource `info://api`, com o mesmo conteúdo Markdown (URL base interpolada dinamicamente).
- Mesmos 2 prompts, com o mesmo texto de instrução para o LLM, incluindo os placeholders de query/name/phone.
- Mesma lógica de negócio: `searchCustomer` filtra em memória (a API legada não tem busca nativa), `id` exato e `name`/`phone` por correspondência parcial case-insensitive.
- Mesmo transporte **stdio** (comunicação via stdin/stdout, JSON-RPC) — usado por Claude Desktop e clientes MCP compatíveis.
- Mesma configuração via variáveis de ambiente: `LEGACY_API_URL` (default `http://localhost:3000`) e `API_SERVICE_TOKEN`.

## O que foi adaptado

| Aspecto | Java | Go |
|---|---|---|
| SDK MCP | Spring AI MCP Server (`spring-ai-mcp-server-spring-boot-starter`) | [SDK oficial Go do MCP](https://github.com/modelcontextprotocol/go-sdk) (`github.com/modelcontextprotocol/go-sdk/mcp`) |
| Definição de tools | Anotações `@Tool`/`@ToolParam` sobre métodos | `mcp.AddTool` genérico com structs `In`/`Out` tipados (schema JSON gerado automaticamente por reflection, igual em espírito às anotações Java) |
| Cliente HTTP para a API legada | `RestClient` (Spring) | `net/http` puro (`client/customer_http_client.go`) |
| **Schema de saída das tools de listagem** | `List<Customer>` retornado diretamente | Envelopado em `{"customers": [...]}` — o SDK Go do MCP exige que o schema de saída estruturada seja do tipo `object` no nível raiz; arrays soltos não são aceitos. O conteúdo textual visto pelo LLM continua sendo o JSON com os mesmos dados. |

## Como executar

```bash
cp .env.example .env   # aponte para a legacy-customer-api-go rodando e informe o service token
go run .
```

Para testar interativamente sem um cliente MCP completo, use o [MCP Inspector](https://modelcontextprotocol.io/docs/tools/inspector):

```bash
npx @modelcontextprotocol/inspector go run .
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Cobrem:
- `CustomerService.SearchCustomer` (filtro por nome parcial case-insensitive, telefone parcial, ID exato, sem filtros) — via um duplo de teste simples no lugar do mock Mockito da versão Java, replicando os mesmos cenários de `CustomerServiceTest.java`.
- `CustomerHttpClient` via `httptest` (listagem, criação, atualização sem campos, remoção com sucesso/erro).
- O servidor MCP **ponta-a-ponta**: um cliente MCP real conectado via `mcp.NewInMemoryTransports()` chamando as 5 tools, lendo o resource e obtendo os 2 prompts — validando o protocolo MCP em si, não apenas a lógica interna.
