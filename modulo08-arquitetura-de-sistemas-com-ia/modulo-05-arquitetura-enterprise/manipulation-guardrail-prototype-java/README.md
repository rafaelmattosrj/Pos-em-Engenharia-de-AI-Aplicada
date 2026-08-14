# Manipulation Guardrail Prototype — Java

Porte Java de **`manipulation-guardrail-prototype.js`** e **`manipulation_guardrail_prototype.py`**
(pasta `..\`, ambos leitura obrigatória como fonte de verdade — os dois são idênticos em
comportamento, esta versão Java replica os dois).

## O que o protótipo faz

Guardrail de manipulação: classifica a pergunta do usuário **antes** de qualquer geração de
resposta, usando o próprio modelo barato (Ollama local, `gemma4:e2b`) como classificador de
segurança. Diferente do sinal de qualidade do Módulo 5.2 (taxa de rejeição no Approval Gate,
medido *depois* que a resposta já foi gerada), este guardrail tenta bloquear *antes* — a resposta
manipulada nunca chega a ser gerada.

Réplica do caso real citado no Módulo 5.2: em janeiro de 2024 um usuário manipulou o chatbot da
DPD a xingar a própria empresa e escrever um poema insultuoso, pedindo para o bot "ignorar
instruções anteriores". Os 4 sinais clássicos de observabilidade (latência, tráfego, erro,
saturação) ficaram todos verdes durante o ataque — nenhum deles pega esse tipo de falha.

O demo roda 3 casos:
1. Pergunta legítima sobre o estudo clínico — deve **passar**.
2. Replay do ataque à DPD, adaptado — deve ser **bloqueado**.
3. Manipulação disfarçada de "auditoria de compliance", sem palavra-gatilho óbvia — deve ser
   **bloqueada** (achado real ao testar o protótipo original: uma primeira versão do prompt do
   classificador listava exemplos de ataque, e um ataque que evitasse essas palavras passava
   direto; esse terceiro caso fica no demo de propósito).

## O que foi mantido 1:1

- O prompt exato do classificador (`INSTRUCAO_CLASSIFICADOR`), incluindo a instrução de responder
  com exatamente uma palavra (`"legitima"` ou `"manipulacao"`) e o teste de "pergunta factual vs.
  pedido de outra coisa", indiferente à autoridade alegada.
- A lógica de decisão: a classificação bruta devolvida pelo modelo é considerada manipulação se
  contiver a substring `"manipul"` (case-insensitive) — mesmo teste do `.includes('manipul')` em
  JS / `"manipul" in classificacao` em Python.
- Os 3 casos de demonstração, com os mesmos textos de pergunta.
- A verificação final: lança erro se o caso 1 for bloqueado (falso positivo) ou se os casos 2/3
  não forem bloqueados (falso negativo).
- Modelo usado (`gemma4:e2b`), sem chave de API — Ollama local.

## O que foi adaptado (e por quê)

- **Chamada ao Ollama**: os originais usam o SDK `ollama-js`/`ollama-python`, que fala com a API
  nativa do Ollama (`POST /api/chat`). Em Java não há um SDK oficial equivalente, então
  `OllamaClassifierClient` chama `POST {OLLAMA_BASE_URL}/api/chat` diretamente via
  `java.net.http.HttpClient` (built-in desde o Java 11, sem dependência extra) + Jackson para
  (de)serializar o JSON. O comportamento observável (mesma pergunta ao mesmo modelo, mesma
  resposta) é idêntico.
- **Sem Spring Boot**: este é um script CLI standalone, sem servidor HTTP — Maven puro é
  suficiente e mantém consistência com os demais protótipos irmãos do módulo. Cogitado usar
  Spring Boot só pela injeção de dependência do `ClassifierClient`, mas o construtor de
  `GuardrailGateway` já cobre isso sem framework.
- **Interface `ClassifierClient`**: extraída para permitir testar `GuardrailGateway` com um dublê
  determinístico (`FakeClassifierClient`), já que os testes automatizados não podem depender de um
  Ollama local rodando durante `mvn test`. O original não tem testes automatizados no sentido
  JUnit — só a asserção final em `main()`, que exige uma classificação real do modelo. Os testes
  Java cobrem a mesma lógica de decisão (e os mesmos 3 cenários) de forma determinística.

## Como rodar

Pré-requisitos: [Ollama](https://ollama.com) instalado e rodando localmente, com o modelo puxado:

```bash
ollama pull gemma4:e2b
ollama serve
```

Compilar e rodar:

```bash
mvn compile
mvn exec:java
```

Configuração opcional via `.env`/variável de ambiente (ver `.env.example`):

```bash
OLLAMA_BASE_URL=http://localhost:11434   # padrão, pode omitir
```

## Como testar

```bash
mvn test
```

Os testes usam `FakeClassifierClient` (em `src/test`) para simular as respostas do classificador
sem depender de um Ollama local — cobrem: parsing da classificação bruta (`isManipulacao`), e os
3 cenários de `main()` (pergunta legítima passa, ataque óbvio bloqueia, ataque disfarçado
bloqueia).
