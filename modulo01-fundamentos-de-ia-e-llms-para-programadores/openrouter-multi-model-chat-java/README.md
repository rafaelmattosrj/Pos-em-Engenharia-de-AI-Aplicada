# OpenRouter Multi-Model Chat com LangChain4j

Demonstração em Java de como conversar com **múltiplos modelos de linguagem** usando o
[OpenRouter](https://openrouter.ai) como gateway unificado, integrado via
[LangChain4j](https://docs.langchain4j.dev).

---

## O que é o OpenRouter?

O **OpenRouter** é um proxy de API que unifica o acesso a dezenas de modelos de linguagem
(Google, Meta, Mistral, OpenAI, Anthropic, etc.) através de **uma única URL base**,
usando o mesmo formato de API do OpenAI. Isso significa que qualquer biblioteca que
suporte a API da OpenAI pode ser redirecionada para o OpenRouter simplesmente trocando
o `baseUrl`.

Vantagens:
- Acesso a modelos gratuitos sem precisar de múltiplas contas.
- Fallback automático entre modelos.
- Dashboard com estatísticas de uso.
- Suporte a streaming, function calling e mais.

---

## Modelos gratuitos demonstrados

| Modelo                                      | Provedor  | Parâmetros |
|---------------------------------------------|-----------|------------|
| `google/gemma-3-27b-it:free`                | Google    | 27B        |
| `meta-llama/llama-3.2-3b-instruct:free`     | Meta      | 3B         |
| `mistralai/mistral-7b-instruct:free`        | Mistral   | 7B         |

> Modelos com o sufixo `:free` não consomem créditos. Você pode explorar todos os modelos
> disponíveis (gratuitos e pagos) em: https://openrouter.ai/models

---

## Pré-requisitos

- Java 17+
- Maven 3.8+
- Uma conta gratuita no OpenRouter e uma chave de API

---

## Como obter sua chave de API

1. Acesse [https://openrouter.ai](https://openrouter.ai) e crie uma conta gratuita.
2. Vá em **Keys** (https://openrouter.ai/keys).
3. Clique em **Create Key** e copie a chave gerada (começa com `sk-or-...`).
4. Você não precisa adicionar créditos para usar modelos gratuitos.

---

## Configuração

1. Copie o arquivo de exemplo de variáveis de ambiente:

   ```bash
   cp .env.example .env
   ```

2. Edite o arquivo `.env` e substitua pelo valor real da sua chave:

   ```env
   OPENROUTER_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   ```

---

## Como executar

```bash
mvn compile exec:java
```

O programa enviará a mesma pergunta para cada um dos três modelos e exibirá as respostas
no terminal:

```
============================================================
  OpenRouter Multi-Model Chat com LangChain4j
============================================================
Pergunta enviada para todos os modelos:
  "Me conte uma curiosidade sobre LLMs (Large Language Models)."
============================================================

Modelo: google/gemma-3-27b-it:free
------------------------------------------------------------
Uma curiosidade fascinante sobre LLMs é...
------------------------------------------------------------

Modelo: meta-llama/llama-3.2-3b-instruct:free
------------------------------------------------------------
...
------------------------------------------------------------

Modelo: mistralai/mistral-7b-instruct:free
------------------------------------------------------------
...
------------------------------------------------------------
```

---

## Estrutura do projeto

```
openrouter-multi-model-chat/
├── src/
│   └── main/
│       └── java/
│           └── com/openrouter/
│               └── Main.java          # Classe principal
├── .env.example                       # Modelo do arquivo de variáveis de ambiente
├── .env                               # Sua chave de API (NÃO commitar!)
├── pom.xml                            # Dependências e configuração Maven
└── README.md
```

---

## Como funciona

O OpenRouter é **compatível com a API da OpenAI**. O LangChain4j fornece a classe
`OpenAiChatModel` que aceita um `baseUrl` customizado. Basta apontá-lo para
`https://openrouter.ai/api/v1` e usar a chave do OpenRouter como `apiKey`:

```java
ChatModel model = OpenAiChatModel.builder()
    .baseUrl("https://openrouter.ai/api/v1")
    .apiKey(System.getenv("OPENROUTER_API_KEY"))
    .modelName("google/gemma-3-27b-it:free")
    .build();

String resposta = model.chat("Me conte uma curiosidade sobre LLMs.");
```

### Nota sobre headers opcionais do OpenRouter

O OpenRouter recomenda o envio de dois headers HTTP opcionais por requisição:

| Header         | Finalidade                                      |
|----------------|-------------------------------------------------|
| `HTTP-Referer` | URL do seu projeto/site                         |
| `X-Title`      | Nome do seu aplicativo                          |

Esses headers aparecem no dashboard de uso do OpenRouter e ajudam a identificar
de onde vêm as requisições. O `OpenAiChatModel` do LangChain4j não expõe headers
HTTP customizados por requisição em sua API de alto nível. Se precisar enviá-los,
use o cliente HTTP diretamente (OkHttp ou `java.net.http.HttpClient`).

---

## Testes

A validação da chave de API (`OpenRouterConfig`) e a orquestração de chamadas a múltiplos
modelos (`MultiModelChatRunner`) foram extraídas de `Main.java` para classes isoladas e
testadas com JUnit 5 + AssertJ — sem depender de chamadas HTTP reais, mesma abordagem usada
no porte Go (`openrouter-multi-model-chat-go/openrouter/client_test.go`, testado com
`httptest`).

Cenários cobertos (equivalentes ao pacote Go `openrouter`):

- resposta com sucesso;
- mensagem de erro da API propagada (ex.: rate limit);
- falha em um modelo sem interromper os demais;
- chave de API ausente/em branco é inválida;
- lista de modelos vazia não realiza chamadas.

```bash
mvn test
```

---

## Dependências principais

| Biblioteca              | Versão   | Finalidade                                  |
|-------------------------|----------|---------------------------------------------|
| `langchain4j`           | 1.0.0    | Core do LangChain4j                         |
| `langchain4j-open-ai`   | 1.0.0    | Integração com APIs compatíveis com OpenAI  |
| `dotenv-java`           | 3.0.0    | Carrega variáveis do arquivo `.env`         |
| `slf4j-simple`          | 2.0.16   | Log em console                              |

---

## Relação com o exemplo shell (Exemplo 11)

Este projeto Java replica o comportamento do script shell que usa `curl` diretamente:

```bash
curl https://openrouter.ai/api/v1/chat/completions \
  -H "Authorization: Bearer $OPENROUTER_API_KEY" \
  -H "HTTP-Referer: https://meusite.com" \
  -H "X-Title: Meu App" \
  -d '{"model": "google/gemma-3-27b-it:free", "messages": [{"role":"user","content":"..."}]}'
```

A diferença é que aqui usamos a abstração do LangChain4j, que gerencia o cliente HTTP,
a serialização JSON e o mapeamento de respostas, além de demonstrar como trocar de
modelo com o mínimo de código repetido.
