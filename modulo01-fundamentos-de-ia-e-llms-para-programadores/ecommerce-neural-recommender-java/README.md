# E-commerce Neural Recommender em Java

Tradução em Java puro do **Exemplo 01** original em JavaScript/TensorFlow.js: um sistema de
recomendação de e-commerce com uma **rede neural feedforward implementada do zero**, sem
nenhum framework de Machine Learning.

## Origem

Porto do exemplo `exemplo-01-ecommerce-recomendations-*` (JavaScript/TensorFlow.js) deste
repositório. A versão em Go (`../ecommerce-neural-recommender-go`) é um porte irmão que segue
a mesma lógica 1:1.

## O que foi mantido 1:1

- Mesmo catálogo de 10 produtos e 5 usuários de demonstração (`data/DataLoader.java`).
- Mesma arquitetura de rede: `[inputDim, 128, 64, 32, 1]`, ReLU nas camadas ocultas, Sigmoid na saída.
- Mesma inicialização de pesos He (`sqrt(2/entradas)`) com seed fixo (`42`), para reprodutibilidade
  dos pesos iniciais.
- Mesmo treino: SGD + backpropagation manual, loss Binary Cross-Entropy, 100 épocas, learning rate
  0.01, shuffle dos exemplos a cada época.
- Mesma codificação de features (pesos `categoria=0.4, cor=0.3, preço=0.2, idade=0.1`) e mesma
  lógica de encoding de usuário (média dos vetores dos produtos comprados, ou apenas idade
  normalizada quando o usuário não tem histórico).
- Mesmos 3 cenários de demonstração no `Main`: recomendações por usuário cadastrado, usuário novo
  sem histórico (id=99, Rafael Souza), usuário com histórico só de acessórios (id=100, Marina Costa).

## O que foi adaptado

- Sem dependências externas de ML — implementação da rede neural (forward/backward/train) em Java
  puro com `java.util.Random`, espelhando a versão Go que usa `math`/`math/rand` da stdlib.
- Estruturas de dados usam `List`/`Map`/records idiomáticos de Java em vez de slices/maps de Go,
  mas a lógica matemática (forward, backward, normalização, encoding) é uma tradução direta.
- `RecommendationEngine.Recommendation` é um `record` Java com `toString()` formatado, equivalente
  à struct de recomendação do Go.

## Como executar

Requer JDK 17+.

```bash
mvn compile exec:java -Dexec.mainClass=com.recomendacao.Main
```

ou, gerando o jar:

```bash
mvn package
java -jar target/ecommerce-recomendations-java-1.0-SNAPSHOT.jar
```

## Como testar

```bash
mvn test
```

Os testes (JUnit 5 + AssertJ) cobrem:

- `DataLoaderTest` — catálogo de produtos/usuários carregado corretamente.
- `ProductUserTest` — modelos de domínio (`Product`, `User`).
- `ModelContextTest` — normalização, cálculo de dimensões do vetor, índices de one-hot encoding,
  limites de idade/preço.
- `NeuralNetworkTest` — forward retorna valor em `[0,1]`, reprodutibilidade da inicialização de
  pesos (seed fixo), redução de loss durante o treino, atualização de pesos via backpropagation.
- `RecommendationEngineTest` — encoding de produto/usuário (com e sem histórico), exceção ao
  chamar `recommend()` antes de treinar, ordenação das recomendações por score, e os mesmos
  cenários de usuário do `Main` (cadastrado, novo sem histórico, histórico de acessórios).

## Credenciais

Este projeto não usa nenhuma API externa nem chave de API — toda a rede neural e os dados de
demonstração são locais, então não há `.env`/`.env.example` a configurar.
