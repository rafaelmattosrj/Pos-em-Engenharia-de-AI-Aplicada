# E-commerce Neural Recommender em Go

Porte em Go do projeto [`ecommerce-neural-recommender-java`](../ecommerce-neural-recommender-java) — sistema de recomendação de e-commerce com rede neural feedforward implementada do zero (sem framework de ML), traduzindo o Exemplo 01 original em JavaScript/TensorFlow.js.

## O que foi mantido 1:1

- Mesmo catálogo de 10 produtos e 5 usuários de demonstração (`data/loader.go`).
- Mesma arquitetura de rede: `[inputDim, 128, 64, 32, 1]`, ReLU nas camadas ocultas, Sigmoid na saída.
- Mesma inicialização de pesos He com seed fixo (42), para reprodutibilidade.
- Mesmo treino: SGD + backpropagation manual, loss Binary Cross-Entropy, 100 épocas, learning rate 0.01, shuffle dos exemplos a cada época.
- Mesma codificação de features (pesos `categoria=0.4, cor=0.3, preço=0.2, idade=0.1`) e mesma lógica de encoding de usuário (média dos vetores dos produtos comprados, ou apenas idade normalizada sem histórico).
- Mesmos 3 cenários de demonstração: recomendações por usuário cadastrado, usuário novo sem histórico (id=99, Rafael Souza), usuário com histórico só de acessórios (id=100, Marina Costa).

## O que foi adaptado

- Sem dependências externas — `math`/`math/rand` da stdlib substituem qualquer biblioteca de ML (nem a versão Java usa uma; ambas implementam a rede do zero).
- Estruturas de dados usam slices/maps idiomáticos de Go em vez de `List`/`Map` do Java, mas a lógica matemática (forward, backward, normalização) é uma tradução direta.

## Como executar

```bash
go run .
```

## Testes

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```
