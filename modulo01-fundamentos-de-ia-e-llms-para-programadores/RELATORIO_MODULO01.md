# Relatório Completo dos Projetos — Módulo 01
**Fundamentos de IA e LLMs para Programadores**
Erick Wendel · Pós-Graduação em Engenharia de Software com IA Aplicada
Rafael Mattos Moreira — Resumo de estudo técnico

---

## PROJETO: exemplo-00-template

### O que faz
Ponto de partida (template) para construir uma rede neural com TensorFlow.js em Node.js. Contém apenas os dados de entrada brutos — 3 pessoas com atributos (idade, cor favorita, localização) e suas categorias (premium, medium, basic) — mas sem o modelo implementado. O aluno deve completar a lógica de treinamento.

### Conceito de IA aplicado
**Rede Neural supervisionada para classificação multiclasse.** Os dados precisam ser convertidos para tensores, normalizados (escala 0-1) e codificados com One-Hot Encoding antes de alimentar a rede. Conceitos: normalização, one-hot encoding, tensor 2D.

### Fluxo de dados
```
Dados brutos (objetos JS)
  → Normalização de idade: (valor - min) / (max - min)
  → One-Hot de cor: [1,0,0] azul | [0,1,0] verde | [0,0,1] vermelho
  → One-Hot de localização: [1,0,0] SP | [0,1,0] RJ | [0,0,1] CTB
  → tf.tensor2d([[...linha de features]])   ← xs (entradas)
  → tf.tensor2d([[1,0,0], [0,1,0], [0,0,1]])  ← ys (labels)
  → [IMPLEMENTAR] trainModel(xs, ys)
  → [IMPLEMENTAR] predict(input)
```

### Arquivos principais
- [index.js](exemplo-00-template/index.js) — dados brutos e estrutura a completar (modelo não implementado)
- [package.json](exemplo-00-template/package.json) — dependência única: `@tensorflow/tfjs-node@4.22`

### Trechos de código importantes
```javascript
// Normalização de idade: transforma 22-30 → 0.0-1.0
// Fórmula: (valor - min) / (max - min)
const idades = [30, 25, 40];
const minIdade = Math.min(...idades);
const maxIdade = Math.max(...idades);
const idadeNorm = (idade) => (idade - minIdade) / (maxIdade - minIdade);

// One-Hot Encoding de cores — cada posição representa uma categoria
// Apenas uma posição tem valor 1 (a categoria ativa)
const cores = { azul: [1, 0, 0], verde: [0, 1, 0], vermelho: [0, 0, 1] };

// Criação de tensor 2D — shape [nAmostras, nFeatures]
const xs = tf.tensor2d([
  [idadeNorm(30), ...cores.azul,   ...localizacoes.SP],  // Erick: premium
  [idadeNorm(25), ...cores.verde,  ...localizacoes.RJ],  // Ana: medium
  [idadeNorm(40), ...cores.vermelho, ...localizacoes.CTB], // Carlos: basic
]);
```

### Dependências e como rodar
```bash
npm install              # instala @tensorflow/tfjs-node
npm start                # node --no-warnings --watch index.js
```
Variáveis de ambiente: **nenhuma**.

### O que eu precisaria mudar para adaptar a um projeto próprio
- Substituir os 3 objetos de pessoas por seu conjunto de dados real
- Mapear suas features para o vetor de entrada (normalizar numéricos, one-hot para categorias)
- Definir as classes de saída (o `ys`) de acordo com o seu problema de classificação
- Implementar as funções `trainModel` e `predict` (ver exemplo-00-z)

### Dúvidas que um dev backend Java/Spring teria

**1. One-Hot Encoding é como uma enum com array? Por que não usar 0, 1, 2?**
Usar 0/1/2 implica ordem numérica (2 > 1 > 0), o que cria um viés errado na rede. One-Hot representa cada categoria como independente — nenhuma tem "mais peso" que outra. Equivale a um `EnumMap<Color, Integer>` onde só uma posição vale 1.

**2. Por que normalizar a idade entre 0 e 1?**
Redes neurais são sensíveis à escala. Se idade vai de 0 a 80 e a cor vai de 0 a 1, o gradiente será dominado pela idade. Normalizar coloca tudo na mesma escala, equivalente a um `@Min(0) @Max(1)` de validação de Bean Validation.

**3. `tf.tensor2d` é como uma matriz bidimensional?**
Exatamente. `shape [3, 9]` significa 3 linhas (amostras) × 9 colunas (features). Equivale a um `double[][]` no Java ou a uma `List<List<Double>>`.

---

## PROJETO: exemplo-00-z

### O que faz
Versão final do exemplo-00-template. Implementa o ciclo completo de treinamento de uma rede neural para classificar usuários em três categorias (premium, medium, basic) com base em idade, cor favorita e localização. Inclui as funções `trainModel` e `predict`, além da predição para um usuário novo ("Zé").

> **Comparação com template**: o template contém apenas os dados brutos; este arquivo adiciona a arquitetura do modelo, treinamento com 100 epochs e a função de predição completa.

### Conceito de IA aplicado
**Rede Neural com classificação multiclasse (Softmax).** Camadas densas com ativação ReLU nas ocultas e Softmax na saída. Loss `categoricalCrossentropy`. Otimizador Adam. O modelo aprende a associar perfis de usuário a categorias de plano.

### Fluxo de dados
```
Dados de treino (xs, ys)
  → model.fit(xs, ys, { epochs: 100, shuffle: true })
     → Camada 1: 80 neurônios, ReLU
     → Camada 2: 3 neurônios, Softmax   ← probabilidades por classe
     → Adam ajusta pesos a cada epoch
  → Dado novo: Zé (normalizado + one-hot)
  → model.predict(input)
     → [0.05, 0.87, 0.08]  ← probabilidade de [premium, medium, basic]
  → argMax → índice 1 → "medium"
```

### Arquivos principais
- [index.js](exemplo-00-z/index.js) — modelo completo com `trainModel()` e `predict()`
- [package.json](exemplo-00-z/package.json) — `@tensorflow/tfjs-node@4.22`

### Trechos de código importantes
```javascript
async function trainModel(xs, ys) {
  const model = tf.sequential();

  // Camada oculta: 80 neurônios com ReLU
  // ReLU zera valores negativos — neurônios "mortos" não propagam erro
  model.add(tf.layers.dense({
    inputShape: [xs.shape[1]], // tamanho automático do vetor de entrada
    units: 80,
    activation: 'relu'
  }));

  // Camada de saída: 3 neurônios (um por classe) com Softmax
  // Softmax converte scores em probabilidades que somam 1
  model.add(tf.layers.dense({
    units: 3,               // premium | medium | basic
    activation: 'softmax'
  }));

  // Adam adapta a taxa de aprendizado — não precisa de lr fixo manual
  // categoricalCrossentropy: loss padrão para classificação multiclasse
  model.compile({
    optimizer: 'adam',
    loss: 'categoricalCrossentropy',
    metrics: ['accuracy']
  });

  await model.fit(xs, ys, {
    epochs: 100,
    shuffle: true  // embaralha a cada epoch para evitar viés de ordem
  });

  return model;
}

async function predict(model, input) {
  // tf.tidy libera memória de tensores intermediários automaticamente
  return tf.tidy(() => {
    const inputTensor = tf.tensor2d([input]);
    const prediction = model.predict(inputTensor);
    return prediction.dataSync(); // converte tensor → Float32Array
  });
}
```

### Dependências e como rodar
```bash
npm install
npm start    # node --no-warnings --watch index.js
```
Variáveis de ambiente: **nenhuma**.

### O que eu precisaria mudar para adaptar a um projeto próprio
- Ajustar `units: 80` da camada oculta (regra geral: entre o tamanho de entrada e saída)
- Mudar o número de neurônios na saída para corresponder ao número de classes do seu problema
- Para classificação binária (sim/não), usar 1 neurônio com `sigmoid` e `binaryCrossentropy`
- Aumentar `epochs` se o modelo não convergir; reduzir se overfitting

### Dúvidas que um dev backend Java/Spring teria

**1. `model.fit()` é como o método `train()` de um `Classifier` do Weka ou Smile?**
Sim, a analogia é direta. `model.fit(xs, ys, { epochs })` é o equivalente a `classifier.buildClassifier(instances)` — passa os dados e o modelo ajusta os pesos internamente. A diferença é que aqui você controla a arquitetura da rede manualmente.

**2. O que é `tf.tidy()` e preciso usar em todo lugar?**
`tf.tidy()` é um escopo de memória para tensores — equivale a um `try-with-resources` do Java. Tensores não são coletados pelo GC normal do JS; `tf.tidy()` descarta todos os tensores intermediários criados dentro do bloco. Use sempre que criar tensores temporários.

**3. `shuffle: true` no `model.fit` tem o mesmo efeito que embaralhar o `DataLoader` do PyTorch?**
Exatamente. Sem shuffle, a rede poderia memorizar a ordem dos dados ao invés dos padrões. É equivalente a `Collections.shuffle(list)` antes de cada epoch de treinamento.

---

## PROJETO: exemplo-01-ecommerce-recomendations-template

### O que faz
Template de aplicação web de e-commerce com sistema de recomendação baseado em TensorFlow.js rodando no navegador. Estrutura MVC completa com Controllers, Services, Views, sistema de Events e Web Workers já scaffoldada — o aluno implementa a lógica de ML de recomendação.

### Conceito de IA aplicado
**Sistema de recomendação colaborativa baseada em perfil de usuário.** Codifica produtos e usuários como vetores numéricos, treina rede neural para prever se um usuário compraria um produto (classificação binária). Conceitos: `encodeProduct`, `encodeUser`, Sigmoid, `binaryCrossentropy`.

### Fluxo de dados
```
users.json + products.json  →  carregados pelos Services
Usuário seleciona perfil   →  UserController dispara evento
Histórico de compras       →  encodeUser() gera vetor de perfil
Cada produto               →  encodeProduct() gera vetor de produto
Par (usuário, produto)     →  concatenados → input da rede
Rede treinada (Web Worker) →  predict() → score 0-1 por produto
Score ordenado             →  top N recomendações exibidas na UI
```

### Arquivos principais
- [src/index.js](exemplo-01-ecommerce-recomendations-template/src/index.js) — bootstrap da aplicação
- [src/controller/ModelTrainingController.js](exemplo-01-ecommerce-recomendations-template/src/controller/ModelTrainingController.js) — orquestra treinamento
- [src/controller/WorkerController.js](exemplo-01-ecommerce-recomendations-template/src/controller/WorkerController.js) — comunicação com Web Worker
- [src/workers/modelTrainingWorker.js](exemplo-01-ecommerce-recomendations-template/src/workers/modelTrainingWorker.js) — treinamento em thread separada
- [src/service/UserService.js](exemplo-01-ecommerce-recomendations-template/src/service/UserService.js) — CRUD de usuários no sessionStorage
- [src/service/ProductService.js](exemplo-01-ecommerce-recomendations-template/src/service/ProductService.js) — carrega produtos do JSON
- [src/events/constants.js](exemplo-01-ecommerce-recomendations-template/src/events/constants.js) — constantes de eventos do sistema
- [data/users.json](exemplo-01-ecommerce-recomendations-template/data/users.json) — 5 usuários com histórico de compras
- [data/products.json](exemplo-01-ecommerce-recomendations-template/data/products.json) — 10 produtos com preço e categoria

### Trechos de código importantes
```javascript
// Sistema de eventos desacoplado — equivalente ao ApplicationEventPublisher do Spring
// Componentes se comunicam via eventos sem referências diretas
const EVENTS = {
  userSelected: 'user:selected',      // usuário trocou de perfil
  purchaseAdded: 'purchase:added',    // produto adicionado ao histórico
  modelTrain: 'training:train',       // dispara treinamento
  trainingComplete: 'training:complete',
  recommendationsReady: 'recommendations:ready',
};

// Web Worker — treinamento em thread separada para não travar a UI
// Equivalente a um @Async do Spring ou CompletableFuture.supplyAsync()
const worker = new Worker('./src/workers/modelTrainingWorker.js', { type: 'module' });
worker.postMessage({ type: EVENTS.modelTrain, data: trainingData });
worker.onmessage = (event) => {
  if (event.data.type === EVENTS.trainingComplete) {
    // modelo pronto — disparar recomendações
  }
};
```

### Dependências e como rodar
```bash
npm install              # instala browser-sync (dev)
npm start                # browser-sync na porta 3000 com live-reload
```
TensorFlow.js é carregado via CDN no `index.html`.
Variáveis de ambiente: **nenhuma**.

### O que eu precisaria mudar para adaptar a um projeto próprio
- Substituir `data/users.json` e `data/products.json` com seus dados reais
- Implementar `encodeProduct()` com as features relevantes ao seu domínio (preço, categoria, atributos)
- Implementar `encodeUser()` como média dos vetores dos produtos no histórico
- Ajustar a arquitetura de camadas no worker para o tamanho do seu vetor de entrada

### Dúvidas que um dev backend Java/Spring teria

**1. Por que Web Worker em vez de simplesmente rodar assíncrono com async/await?**
`async/await` não cria threads — ainda roda na thread principal do JS. O treinamento de ML é CPU-intensivo e bloquearia a UI por segundos. Web Worker é como um `Thread` do Java: executa em paralelo real, sem compartilhar memória diretamente.

**2. O sessionStorage é como HttpSession do Servlet? Os dados somem ao fechar o browser?**
Exato. `sessionStorage` persiste apenas enquanto a aba estiver aberta — equivalente ao escopo de sessão do Spring (`@SessionScope`). Ao fechar, os dados são perdidos. Para persistência maior, usaria `localStorage`.

**3. O sistema de eventos é equivalente ao padrão Observer/EventBus do Spring?**
Sim. `dispatchEvent(new CustomEvent(type, { detail: data }))` + `addEventListener(type, handler)` é o equivalente ao `ApplicationEventPublisher.publishEvent()` + `@EventListener` do Spring, mas no DOM.

---

## PROJETO: exemplo-01-ecommerce-recomendations-z

### O que faz
Versão final do sistema de recomendação de e-commerce com TensorFlow.js, dividida em **5 partes iterativas** que mostram a evolução da implementação. Cada parte adiciona uma camada de funcionalidade sobre a anterior, do encoding básico até a recomendação completa com visualização via tfvis.

> **Comparação com template**: o template tem a estrutura MVC pronta mas sem ML implementado. As 5 partes do -z mostram a implementação progressiva: encoding → modelo → worker → recomendação → visualização.

### Conceito de IA aplicado
**Sistema de recomendação binária (comprou/não comprou).** Rede 128→64→32→1 neurônios com Sigmoid na saída. Loss `binaryCrossentropy`. Para cada par (usuário, produto), a rede retorna a probabilidade de compra.

### Fluxo de dados
```
Parte 01: encodeProduct() + encodeUser() → vetores numéricos
Parte 02: modelo sequencial 128→64→32→1 (Sigmoid) compilado
Parte 03: treinamento no Web Worker com postMessage/onmessage
Parte 04: predict() para todos os produtos → scores ordenados
Parte 05: tfvis → gráficos de loss/accuracy durante treinamento
```

### Arquivos principais
Cada subpasta (`parte01` a `parte05`) contém estrutura idêntica ao template.
- [parte01/.../modelTrainingWorker.js](exemplo-01-ecommerce-recomendations-z/parte01-ecommerce-recomendations-with-tensorflow/src/workers/modelTrainingWorker.js) — encoding implementado
- [parte03/.../WorkerController.js](exemplo-01-ecommerce-recomendations-z/parte03-ecommerce-recomendations-with-tensorflow/src/controller/WorkerController.js) — comunicação Worker completa
- [parte05/.../TFVisorController.js](exemplo-01-ecommerce-recomendations-z/parte05-ecommerce-recomendations-with-tensorflow/src/controller/TFVisorController.js) — visualização com tfvis

### Trechos de código importantes
```javascript
// Arquitetura da rede de recomendação (parte02)
// 128 → 64 → 32 neurônios: padrão funil — extrai features progressivamente
const model = tf.sequential({
  layers: [
    tf.layers.dense({ inputShape: [inputSize], units: 128, activation: 'relu' }),
    tf.layers.dense({ units: 64, activation: 'relu' }),
    tf.layers.dense({ units: 32, activation: 'relu' }),
    tf.layers.dense({ units: 1, activation: 'sigmoid' }), // 0-1: probabilidade de compra
  ]
});

model.compile({
  optimizer: tf.train.adam(0.01),  // learning rate explícita: 0.01
  loss: 'binaryCrossentropy',      // loss para classificação binária
  metrics: ['accuracy']
});

// encodeProduct: normaliza preço e codifica categoria/cor com one-hot
// O vetor final representa o "DNA numérico" do produto
function encodeProduct(product, allProducts) {
  const prices = allProducts.map(p => p.price);
  const priceNorm = (product.price - Math.min(...prices)) / (Math.max(...prices) - Math.min(...prices));
  const categoryOneHot = categories.map(c => c === product.category ? 1 : 0);
  return [priceNorm, ...categoryOneHot];
}

// encodeUser: média dos vetores de todos os produtos já comprados
// Cria um "centróide" do perfil de consumo do usuário
function encodeUser(user, allProducts) {
  const purchasedVectors = user.purchases.map(pid => {
    const product = allProducts.find(p => p.id === pid);
    return encodeProduct(product, allProducts);
  });
  // média coluna a coluna → perfil médio do usuário
  return purchasedVectors[0].map((_, i) =>
    purchasedVectors.reduce((sum, v) => sum + v[i], 0) / purchasedVectors.length
  );
}
```

### Dependências e como rodar
```bash
cd parte05-ecommerce-recomendations-with-tensorflow
npm install
npm start    # porta 3000
```
Variáveis de ambiente: **nenhuma**.

### O que eu precisaria mudar para adaptar a um projeto próprio
- Ajustar `encodeProduct` com as features do seu catálogo (mais categorias, mais atributos)
- Aumentar a camada de entrada (inputShape) proporcionalmente ao vetor de features
- Substituir os dados estáticos de JSON por chamadas a uma API REST real
- Em produção, mover o treinamento para o servidor (Node.js ou Python) e servir apenas o modelo treinado

### Dúvidas que um dev backend Java/Spring teria

**1. Sigmoid retorna 0 a 1 — mas como interpreto isso como "vai comprar"?**
Define-se um threshold (ex: 0.5). Score > 0.5 = vai comprar. É como um `if (probabilidade > 0.5) return true`. O treinamento ajusta os pesos para que compras reais gerem scores altos e não-compras gerem scores baixos.

**2. A arquitetura 128→64→32 foi escolhida como?**
Convenção empírica: diminuir gradualmente (funil) extrai representações mais abstratas a cada camada. Não há fórmula exata — é hiperparâmetro. Começa com uma arquitetura razoável, avalia a loss e ajusta. Equivale a escolher tamanho de pool de threads sem métricas.

**3. `binaryCrossentropy` vs `categoricalCrossentropy` — quando usar cada um?**
`binary`: 2 classes (comprou/não comprou) com 1 neurônio Sigmoid na saída. `categorical`: 3+ classes com N neurônios Softmax. É a diferença entre um `boolean` e um `enum` na saída do seu classificador.

---

## PROJETO: exemplo-02-vencendo-qualquer-jogo

### O que faz
Implementação do jogo **DuckHunt** em JavaScript com automação via **YOLO** (modelo de detecção de objetos). O ML detecta patos na tela em tempo real e simula cliques automaticamente nas posições calculadas, demonstrando IA aplicada a jogos sem acesso ao código interno do jogo.

### Conceito de IA aplicado
**YOLO (You Only Look Once)** — modelo de detecção de objetos em imagens. Analisa cada frame do jogo, identifica os patos (bounding boxes com coordenadas) e calcula o centro para simular o clique. Processado em **Web Worker** para não travar a UI.

### Fluxo de dados
```
Frame do jogo (canvas)
  → createImageBitmap()              ← captura frame atual
  → tf.browser.fromPixels(bitmap)   ← converte para tensor
  → tf.image.resizeBilinear([640,640]) ← YOLO espera 640x640
  → divisão por 255 (normalização 0-1)
  → tensor.expandDims(0)             ← adiciona dim de batch [1,640,640,3]
  → model.executeAsync(input)        ← inferência YOLO
  → boxes, scores, classes           ← saídas do modelo
  → filtrar score > 40% e class === "bird"
  → centerX = x1 + (x2-x1)/2
  → centerY = y1 + (y2-y1)/2
  → simulateClick(centerX, centerY)  ← abate o pato
```

### Arquivos principais
- [DuckHunt-JS-parte01/src/machine-learning/main.js](exemplo-02-vencendo-qualquer-jogo/DuckHunt-JS-parte01/src/machine-learning/main.js) — pipeline YOLO completo
- [DuckHunt-JS-parte01/src/modules/Game.js](exemplo-02-vencendo-qualquer-jogo/DuckHunt-JS-parte01/src/modules/Game.js) — lógica principal do jogo (647 linhas, PixiJS)
- [DuckHunt-JS-parte01/package.json](exemplo-02-vencendo-qualquer-jogo/DuckHunt-JS-parte01/package.json) — PixiJS, Webpack, Babel, Howler.js, GSAP
- [DuckHunt-JS-parte01/main.js](exemplo-02-vencendo-qualquer-jogo/DuckHunt-JS-parte01/main.js) — entry point que integra Game + ML

### Trechos de código importantes
```javascript
// main.js — entry point que integra o jogo com o modelo YOLO
import main from './machine-learning/main';  // pipeline de detecção
import Game from './src/modules/Game';

document.addEventListener('DOMContentLoaded', async function () {
  const game = new Game({ spritesheet: 'sprites.json' });
  await game.load();
  await main(game);  // passa instância do jogo para o ML controlar
});

// machine-learning/main.js — pipeline YOLO
async function detectDucks(frame, model) {
  // 1. Pré-processamento: resize + normalização
  const input = tf.tidy(() => {
    return tf.browser.fromPixels(frame)
      .resizeBilinear([640, 640])  // YOLO requer 640x640
      .div(255.0)                  // normaliza pixels para 0-1
      .expandDims(0);              // [H,W,C] → [1,H,W,C] (batch dimension)
  });

  // 2. Inferência — model.executeAsync para modelos com múltiplas saídas
  const [boxes, scores, classes] = await model.executeAsync(input);

  // 3. Filtro por confiança e classe alvo
  const scoresData = await scores.data();
  const boxesData = await boxes.data();

  const ducks = [];
  for (let i = 0; i < scoresData.length; i++) {
    if (scoresData[i] > 0.4 && classes[i] === DUCK_CLASS_ID) {
      const [y1, x1, y2, x2] = boxesData.slice(i * 4, i * 4 + 4);
      ducks.push({
        centerX: x1 + (x2 - x1) / 2,  // centro horizontal da bounding box
        centerY: y1 + (y2 - y1) / 2,  // centro vertical da bounding box
        score: scoresData[i]
      });
    }
  }
  return ducks;
}

// Warm-up: executa predição com tensor de zeros para "aquecer" o modelo
// Melhora performance das predições subsequentes (cache interno do TF)
await model.executeAsync(tf.zeros([1, 640, 640, 3]));
```

### Dependências e como rodar
```bash
npm install
npm start        # webpack-dev-server na porta 8080
npm run build    # compila para produção
```
Variáveis de ambiente: **nenhuma**.

### O que eu precisaria mudar para adaptar a um projeto próprio
- Substituir o modelo YOLO por outro treinado na sua classe de objetos alvo (ex: detectar rostos, carros, QR codes)
- Ajustar `DUCK_CLASS_ID` para o índice da sua classe no modelo utilizado
- Alterar o threshold de `0.4` (40%) para mais alto se houver falsos positivos
- O pipeline de pré-processamento (640×640, normalização) é padrão YOLO — manter igual

### Dúvidas que um dev backend Java/Spring teria

**1. `model.executeAsync` vs `model.predict` — qual usar?**
`model.predict` é síncrono e para modelos simples com uma saída. `model.executeAsync` é assíncrono e necessário para modelos com múltiplas saídas (como YOLO que retorna boxes + scores + classes). Prefira `executeAsync` para modelos importados de PyTorch/TF2.

**2. Por que o warm-up com tensor de zeros?**
Na primeira inferência, o modelo carrega pesos na memória da GPU/CPU e compila kernels internos — isso causa latência alta (~500ms). O warm-up faz esse custo antecipadamente, em background, sem impactar o usuário. Equivalente a "pré-carregar" um Spring Bean singleton na inicialização.

**3. PixiJS é como a camada de rendering — é necessário conhecê-lo para o ML?**
Não. O PixiJS cuida apenas do rendering do jogo (sprites, canvas, animações). O ML é independente — recebe apenas o frame atual como bitmap. Você poderia trocar o jogo por qualquer outra fonte de imagem.

---

## PROJETO: exemplo-03-webai01

### O que faz
Demonstração mínima da **Web AI API** nativa do Chrome — utiliza o modelo Gemini Nano que roda localmente na máquina do usuário, sem servidor externo. Faz uma pergunta fixa ("Quem inventou o JavaScript?") e exibe a resposta em streaming com renderização Markdown.

### Conceito de IA aplicado
**LLM nativa no navegador (Web 4.0).** O Chrome disponibiliza o modelo Gemini Nano via API experimental `LanguageModel`. Streaming de tokens — a resposta aparece palavra por palavra, sem esperar o texto completo. O modelo roda 100% localmente, sem custo de API.

### Fluxo de dados
```
LanguageModel.params()              ← lê parâmetros padrão do modelo local
  → { defaultTemperature, defaultTopK }
LanguageModel.create({ initialPrompts })  ← inicializa sessão com system prompt
  → session
session.promptStreaming([{ role:'user', content: question }])
  → AsyncIterator de tokens
  → fullText += token por token
  → markdown.toHTML(fullText) → output.innerHTML  ← renderiza ao vivo
```

### Arquivos principais
- [index.html](exemplo-03-webai01/index.html) — arquivo único com todo o código embutido na tag `<script>`

### Trechos de código importantes
```javascript
// Lê parâmetros do modelo instalado no Chrome (temperatura e top-k)
const params = await LanguageModel.params();
const { defaultTemperature, defaultTopK } = params;

// System prompt — equivalente ao "role: system" das APIs de LLM
const initialPrompts = [{
  role: 'system',
  content: 'Você é um assistente de IA que responde de forma clara e objetiva.'
}];

// Cria sessão com o modelo local — pode precisar baixar o modelo (~2.5GB) na 1a vez
const session = await LanguageModel.create({
  expectedInputLanguages: ["pt"],
  temperature: defaultTemperature,
  topK: defaultTopK,
  initialPrompts,
});

// Streaming: for-await itera token por token conforme a IA gera
// Cada iteração: token = próximo trecho de texto gerado
const responseStream = await session.promptStreaming([{
  role: 'user',
  content: 'Quem inventou o JavaScript?'
}]);

let fullText = "";
for await (const token of responseStream) {
  fullText += token;
  output.innerHTML = markdown.toHTML(fullText); // re-renderiza markdown a cada token
}
```

### Dependências e como rodar
Não há `npm install` — tudo via CDN.
```
Requisito: Chrome Canary ou Chrome com flags ativadas:
  chrome://flags/#prompt-api-for-gemini-nano → Enabled
  chrome://flags/#optimization-guide-on-device-model → Enabled BypassPerfRequirement
```
Abrir `index.html` diretamente no Chrome (ou via live server).

Variáveis de ambiente: **nenhuma**.

### O que eu precisaria mudar para adaptar a um projeto próprio
- Alterar o `system prompt` para o contexto do seu assistente
- Substituir a pergunta fixa por input dinâmico do usuário
- Adicionar persistência do histórico (array de mensagens) para conversa multi-turn
- Verificar se o usuário tem o modelo instalado antes de usar (tratar `LanguageModel.params()` fallback)

### Dúvidas que um dev backend Java/Spring teria

**1. É seguro usar essa API em produção?**
Não ainda. É experimental e disponível apenas no Chrome (não Firefox, Safari, Edge). Ideal para prototipagem e demos. Para produção, use API de LLM com SLA (OpenAI, Anthropic, OpenRouter).

**2. O `for await` é como um `InputStream.read()` do Java?**
Boa analogia. É um `AsyncIterator` — cada iteração aguarda o próximo token gerado. Equivale a ler um `BufferedReader` linha por linha de um `StreamingResponseBody` do Spring.

**3. Por que o modelo precisa de download (~2.5GB)? Onde fica armazenado?**
O Chrome gerencia o modelo localmente no sistema de arquivos do usuário (similar a um cache do browser). Uma vez baixado, as requisições subsequentes são instantâneas. Isso é o conceito de "Web 4.0" — processamento de IA sem depender de servidores.

---

## PROJETO: exemplo-04-webai02-temperature-and-topK

### O que faz
Interface web completa para experimentar os parâmetros **Temperature** e **Top-K** do modelo de IA nativo do Chrome em tempo real. O usuário ajusta sliders, digita uma pergunta e cancela a geração a qualquer momento com AbortController.

### Conceito de IA aplicado
**Parâmetros de sampling de LLMs.** Temperature controla a aleatoriedade da geração (valores altos = respostas mais criativas/imprevisíveis). Top-K limita o vocabulário às K palavras mais prováveis a cada passo. Ambos afetam diretamente a "personalidade" das respostas do modelo.

### Fluxo de dados
```
Usuário ajusta Temperature (slider) + Top-K (input)
  → onSubmitQuestion()
  → LanguageModel.create({ temperature, topK })  ← parâmetros customizados
  → session.promptStreaming(messages)
  → for await (token)
  → exibe resposta progressivamente
  → [cancelar] abortController.abort()
     → interrompe o stream imediatamente
```

### Arquivos principais
- [index.html](exemplo-04-webai02-temperature-and-topK/index.html) — UI com sliders de Temperature e Top-K
- [index.js](exemplo-04-webai02-temperature-and-topK/index.js) — lógica completa (222 linhas): sessão, streaming, AbortController
- [style.css](exemplo-04-webai02-temperature-and-topK/style.css) — tema escuro com sliders customizados (227 linhas)
- [package.json](exemplo-04-webai02-temperature-and-topK/package.json) — só `http-server` como devDependency

### Trechos de código importantes
```javascript
// Contexto de estado da IA — evita múltiplas sessões simultâneas
const aiContext = {
  session: null,
  abortController: null,
  isGenerating: false,
};

// Generator que cria sessão com parâmetros customizados do usuário
async function* askAI(question, temperature, topK) {
  aiContext.abortController = new AbortController();

  // Cria nova sessão com os parâmetros escolhidos pelo usuário no slider
  aiContext.session = await LanguageModel.create({
    expectedInputLanguages: ["pt"],
    temperature: parseFloat(temperature),  // ex: 0.8 = mais criativo
    topK: parseInt(topK),                  // ex: 40 = limita a 40 tokens possíveis
    initialPrompts: [{
      role: 'system',
      content: 'Você é um assistente de IA que responde de forma clara e objetiva.'
    }]
  });

  const stream = await aiContext.session.promptStreaming(
    [{ role: 'user', content: question }],
    { signal: aiContext.abortController.signal }  // cancela com abort()
  );

  for await (const token of stream) {
    yield token;  // propaga token por token para o chamador
  }
}

// Botão "Parar" — cancela a geração imediatamente
function stopGeneration() {
  aiContext.abortController?.abort();  // sinaliza cancelamento
  aiContext.isGenerating = false;
  toggleSendOrStopButton(false);
}
```

### Dependências e como rodar
```bash
npm install         # instala http-server
npm start           # http-server na porta padrão 8080
```
Mesmo requisito do exemplo-03: Chrome com flags de IA ativadas.
Variáveis de ambiente: **nenhuma**.

### O que eu precisaria mudar para adaptar a um projeto próprio
- Adicionar persist dos parâmetros no `localStorage` para o usuário não reconfigurar toda vez
- Expor o `temperature` e `topK` via querystring para compartilhar configurações
- Adicionar histórico de conversas (multi-turn) com array de mensagens acumulado
- Para produção sem Web AI API: trocar para OpenAI/Anthropic mantendo a mesma interface UI

### Dúvidas que um dev backend Java/Spring teria

**1. `AbortController` é equivalente ao `CancellationToken` do .NET ou `Future.cancel()` do Java?**
Sim. `AbortController.abort()` sinaliza cancelamento ao stream. No Spring, seria `DeferredResult.setResult()` com um valor de cancelamento ou uso do `@RequestMapping` com `Callable` + `TimeoutValueCallable`. O `signal` é passado ao stream e ele para quando detecta o sinal.

**2. Temperature 0 vs Temperature 1 — qual usar em produção?**
Temperature baixa (~0.2-0.4): respostas determinísticas, factuais, boas para Q&A com dados reais. Temperature alta (~0.8-1.0): respostas criativas, variadas, boas para brainstorming. Para sistemas de suporte ao cliente, use 0.3. Para geração de marketing, use 0.7-0.9.

**3. Top-K e Top-P (nucleus sampling) — qual a diferença prática?**
Top-K limita a escolha aos K tokens mais prováveis (quantidade fixa). Top-P soma probabilidades até atingir P% e escolhe dentro desse subconjunto (quantidade variável). Top-P é mais flexível — quando o modelo está confiante, considera menos opções; quando incerto, considera mais.

---

## PROJETO: exemplo-05-webai03-multimodal

### O que faz
Versão avançada da Web AI com suporte **multimodal** — o usuário envia texto, imagem ou áudio junto com a pergunta. Inclui tradução automática para português via Translator API e detecção de idioma. Arquitetura MVC separada em Services, Controller e View.

### Conceito de IA aplicado
**Multimodalidade** — a IA processa texto + imagem + áudio simultaneamente. Usa três APIs experimentais do Chrome: `LanguageModel` (LLM), `Translator` (tradução) e `LanguageDetector` (detecção de idioma). Arquitetura MVC com separação de responsabilidades.

### Fluxo de dados
```
Usuário: pergunta (texto) + arquivo opcional (imagem/áudio)
  → AIService.checkRequirements()     ← valida Chrome + 3 APIs disponíveis
  → TranslationService.initialize()  ← baixa modelo de tradução local
  → AIService.createSession(question, file, temperature, topK)
     → se arquivo: anexa ao prompt como parte multimodal
     → LanguageModel.create() → session
     → session.promptStreaming()
     → yield token por token
  → FormController recebe tokens → View.appendText()
  → resposta completa → TranslationService.translate(response, 'pt')
  → View exibe resposta traduzida
```

### Arquivos principais
- [services/aiService.js](exemplo-05-webai03-multimodal/services/aiService.js) — LanguageModel API, suporte a texto/imagem/áudio (172 linhas)
- [services/translationService.js](exemplo-05-webai03-multimodal/services/translationService.js) — Translator API + LanguageDetector API
- [controllers/formController.js](exemplo-05-webai03-multimodal/controllers/formController.js) — orquestra AI + Translation + View (108 linhas)
- [views/view.js](exemplo-05-webai03-multimodal/views/view.js) — manipulação do DOM
- [index.js](exemplo-05-webai03-multimodal/index.js) — bootstrap MVC (40 linhas)
- [index.html](exemplo-05-webai03-multimodal/index.html) — UI com upload de arquivo

### Trechos de código importantes
```javascript
// index.js — bootstrap limpo com injeção de dependências manual
// Equivalente ao ApplicationContext do Spring configurando os Beans
import { AIService } from './services/aiService.js';
import { TranslationService } from './services/translationService.js';
import { View } from './views/view.js';
import { FormController } from './controllers/formController.js';

(async function main() {
  const aiService = new AIService();
  const translationService = new TranslationService();
  const view = new View();

  // Verifica se o browser suporta as APIs necessárias
  const errors = await aiService.checkRequirements();
  if (errors) { view.showError(errors); return; }

  // Inicializa tradução local (baixa modelo se necessário)
  await translationService.initialize();

  // Wiring manual dos componentes — Controller recebe todas as dependências
  const controller = new FormController(aiService, translationService, view);
  controller.setupEventListeners();
})();

// aiService.js — generator com suporte a 3 tipos de mídia
async function* createSession(question, file, temperature, topK) {
  const session = await LanguageModel.create({ temperature, topK, initialPrompts });

  // Monta prompt diferente dependendo do tipo de arquivo
  let prompt;
  if (!file) {
    prompt = [{ role: 'user', content: question }];
  } else if (file.type.startsWith('image/')) {
    prompt = [{ role: 'user', content: [
      { type: 'image', image: file },    // imagem como parte do prompt
      { type: 'text',  text: question }
    ]}];
  } else {
    prompt = [{ role: 'user', content: [
      { type: 'audio', audio: file },    // áudio como parte do prompt
      { type: 'text',  text: question }
    ]}];
  }

  const stream = await session.promptStreaming(prompt, { signal: this.abortController.signal });
  for await (const token of stream) yield token;
}
```

### Dependências e como rodar
```bash
npm install         # http-server
npm start           # porta 8080
```
Requer Chrome com flags ativadas para: LanguageModel, Translator, LanguageDetector.
Variáveis de ambiente: **nenhuma**.

### O que eu precisaria mudar para adaptar a um projeto próprio
- Usar o padrão MVC desta estrutura como base para qualquer app de IA no browser
- Adicionar suporte a PDF (converter para imagem antes de enviar)
- Implementar histórico de conversa persistente com `localStorage`
- Para produção: trocar `AIService` por uma implementação que chama API REST (OpenAI/Anthropic), mantendo o mesmo contrato das interfaces

### Dúvidas que um dev backend Java/Spring teria

**1. `(async function main() { ... })()` — por que esta sintaxe?**
É uma IIFE (Immediately Invoked Function Expression) assíncrona. Em módulos ES6, o `await` no nível superior já é suportado, mas essa sintaxe garante compatibilidade e deixa claro que é o ponto de entrada. Equivalente ao `public static void main(String[] args)`.

**2. A injeção de dependências manual (sem framework) funciona bem em JS?**
Para projetos pequenos, sim. O problema escala quando há muitas dependências — aí surgem frameworks como Angular DI ou bibliotecas como Inversify. Para esse porte de projeto, o wiring manual no `main()` é mais legível que annotations.

**3. Como verificar se o usuário tem os modelos baixados antes de exibir a UI?**
`await aiService.checkRequirements()` já faz isso — verifica se as APIs existem e se o modelo está disponível. Se `LanguageModel.availability()` retornar `'no'`, o modelo não está instalado. Se retornar `'after-download'`, precisa baixar primeiro (exibir progress bar).

---

## PROJETO: exemplo-06-playwright-testes

### O que faz
Conjunto de **prompts e configuração MCP** para usar o Playwright MCP no VS Code/Cursor e gerar testes End-to-End automatizados com IA. A IA navega no site, observa o comportamento real e gera testes TypeScript válidos que passam na primeira execução.

### Conceito de IA aplicado
**MCP (Model Context Protocol)** — protocolo da Anthropic que conecta a IA a ferramentas externas. O Playwright MCP dá à IA controle de um browser real. A IA usa o que observa para gerar testes idempotentes sem "inventar" seletores.

### Fluxo de dados
```
Desenvolvedor abre VS Code/Cursor com MCP configurado
  → usa prompt de generate_test.prompt.md
  → IA recebe instrução + ferramentas Playwright disponíveis
  → IA navega no site passo a passo (screenshot, click, fill)
  → IA observa o DOM e comportamentos reais
  → IA gera teste TypeScript baseado no que viu (não no que imaginou)
  → IA salva em tests/ e executa
  → se falhar → IA itera e corrige automaticamente
```

### Arquivos principais
- [example.mcp.json](exemplo-06-playwright-testes/example.mcp.json) — configuração do servidor MCP Playwright
- [prompts/generate_test.prompt.md](exemplo-06-playwright-testes/prompts/generate_test.prompt.md) — instrução para a IA gerar testes
- [prompts/generate-tests.md](exemplo-06-playwright-testes/prompts/generate-tests.md) — prompt simples com URL e cenários
- [prompts/project-scaffolding.md](exemplo-06-playwright-testes/prompts/project-scaffolding.md) — setup completo com CI/GitHub Actions

### Trechos de código importantes
```json
// example.mcp.json — configura o servidor MCP no editor
// Equivalente a um plugin que dá à IA controle do browser
{
  "servers": {
    "playwright": {
      "command": "npx",
      "args": ["@playwright/mcp@latest", "--extension"],
      "env": {
        "PLAYWRIGHT_MCP_EXTENSION_TOKEN": "YOUR_TOKEN_HERE"
      }
    }
  }
}
```

```markdown
// generate_test.prompt.md — instrução para a IA (regras do agente)
// Regras críticas que controlam o comportamento:
- DO NOT generate test code based on the scenario alone.    ← sem "inventar"
- DO run steps one by one using the tools provided.        ← executa antes de gerar
- Only after all steps are completed, emit a Playwright TypeScript test
- Execute the test file and iterate until the test passes  ← auto-correção
- Use Chrome for testing instead of headless browsers
- Keep tests idempotent                                    ← sem dependência de estado
- Prefer getByRole + names over brittle selectors          ← seletores semânticos
```

### Dependências e como rodar
```bash
# 1. Instalar extensão Playwright MCP no VS Code
# 2. Copiar example.mcp.json para .mcp.json na raiz do projeto
# 3. Obter token em: https://playwright.dev/docs/mcp

# 4. No editor com Cursor/VS Code Agent, usar o prompt:
# "use o prompt em prompts/generate_test.prompt.md e navegue em https://..."

# 5. Para scaffolding completo (package.json + CI):
# Copiar conteúdo de project-scaffolding.md e colar no chat do agente
```
Variáveis de ambiente: `PLAYWRIGHT_MCP_EXTENSION_TOKEN`

### O que eu precisaria mudar para adaptar a um projeto próprio
- Atualizar a URL no `generate-tests.md` para a sua aplicação
- Adicionar cenários específicos do seu domínio (ex: login, checkout, cadastro)
- Configurar `baseURL` no `playwright.config.ts` gerado
- Para CI: ajustar o workflow GitHub Actions gerado pelo `project-scaffolding.md`

### Dúvidas que um dev backend Java/Spring teria

**1. Isso substitui escrever testes manuais com Playwright/Selenium?**
Substitui a fase de escrita inicial. A IA gera o teste, mas você precisa validar, manter e refatorar. É como ter um junior que escreve o primeiro rascunho — você revisa e aprova. Para cenários complexos, ainda é mais rápido escrever manualmente.

**2. `getByRole` vs `getElementById` — por que preferir roles?**
`getElementById("submit-btn")` quebra se o ID mudar. `getByRole('button', { name: 'Enviar' })` busca pelo papel semântico + texto visível, resistente a mudanças de CSS/ID. Equivale a testar comportamento, não implementação — princípio do Testing Library.

**3. "Idempotente" nos testes significa o quê na prática?**
O teste pode rodar N vezes em qualquer ordem e produzir o mesmo resultado. Não depende de dados criados por outro teste, não deixa lixo no banco. No Spring, equivale a usar `@Transactional` nos testes para rollback automático após cada execução.

---

## PROJETO: exemplo-07-playwright-navegacao

### O que faz
Demonstração de **automação web com IA** via Playwright MCP. O agente navega no Sessionize (perfil público de palestrante), extrai dados automaticamente e preenche um formulário Google Forms com as informações coletadas — tudo através de um único prompt em linguagem natural.

### Conceito de IA aplicado
**Agente de IA com ferramentas de navegação (MCP Playwright).** A IA recebe um objetivo em linguagem natural, planeja os passos, navega em múltiplos sites, extrai informações contextuais e executa ações — sem código explícito para cada passo.

### Fluxo de dados
```
Prompt em linguagem natural (prompt.md)
  → IA interpreta objetivo: coletar dados + preencher formulário
  → Playwright MCP: navigate(forms.gle URL)
  → IA identifica campos obrigatórios do formulário
  → Playwright MCP: navigate(sessionize.com/erickwendel)
  → IA extrai: nome, bio, redes sociais, palestra em PT com "javascript"
  → Playwright MCP: navigate(formulário de volta)
  → IA preenche cada campo com dados extraídos
  → IA para antes do submit (conforme instrução)
```

### Arquivos principais
- [prompt.md](exemplo-07-playwright-navegacao/prompt.md) — instrução completa em português para o agente
- [example.mcp.json](exemplo-07-playwright-navegacao/example.mcp.json) — configuração MCP idêntica ao exemplo-06

### Trechos de código importantes
```markdown
// prompt.md — prompt completo do agente de navegação
// Demonstra como instruções em linguagem natural controlam comportamento complexo:

"Navegue até o formulário https://forms.gle/... e veja quais campos são necessários.

Então navegue até a página do palestrante em https://sessionize.com/erickwendel,
obtenha todos os dados do perfil que o formulário pede a partir desta página
e então escolha uma palestra em português que tenha javascript no título
e preencha o formulário.

Não aperte o botão submit pois quero validar o processo.
Garanta que todas as informações são em português."

// Princípios de prompt para agentes:
// 1. Objetivo claro e mensurável
// 2. Ordem das ações especificada
// 3. Critérios de filtragem explícitos ("javascript no título", "em português")
// 4. Restrições de segurança ("não aperte submit")
```

### Dependências e como rodar
Mesma configuração do exemplo-06 (Playwright MCP).
```bash
# Usar o prompt.md no chat do agente (Cursor/VS Code com MCP configurado)
```
Variáveis de ambiente: `PLAYWRIGHT_MCP_EXTENSION_TOKEN`

### O que eu precisaria mudar para adaptar a um projeto próprio
- Trocar as URLs para os seus sistemas (formulários internos, portais corporativos)
- Ajustar os critérios de filtragem para o seu domínio
- Adicionar validação dos dados antes do preenchimento
- Para automação recorrente: transformar o prompt em um script Node.js usando `@playwright/test`

### Dúvidas que um dev backend Java/Spring teria

**1. Isso é mais poderoso que Selenium tradicional? Quando usar cada um?**
Para scraping e automação ad-hoc com sites variados: IA + MCP é mais rápido. Para testes de regressão estáveis e pipeline CI/CD: Playwright/Selenium com código explícito é mais confiável. Use IA para exploração; use código para automação de produção.

**2. O agente pode preencher qualquer formulário ou só os que conhece?**
Qualquer formulário visível no browser, porque ele "vê" o DOM como um humano vê a tela. Não precisa conhecer o formulário previamente — navega, lê os labels, preenche. Limitação: formulários com CAPTCHA ou anti-bot.

**3. Como garantir que a IA não clicará em "submit" por engano?**
Instruindo explicitamente no prompt ("Não aperte o botão submit"). A IA tende a obedecer restrições claras. Para mais segurança: usar um ambiente de staging, não produção, para experimentações com agentes.

---

## PROJETO: exemplo-08-context7

### O que faz
Projeto **Next.js com autenticação GitHub OAuth** gerado pelo agente de IA usando o **MCP Context7** (documentação atualizada de frameworks) + **MCP Playwright** (validação). Demonstra como Context7 injeta docs reais no contexto da IA para evitar código com APIs desatualizadas.

> **Contexto do exemplo**: o prompt mostra como pedir ao agente que use Context7 para buscar a documentação atual do Better Auth antes de gerar código — garantindo que a implementação use a versão mais recente da biblioteca.

### Conceito de IA aplicado
**MCP Context7** — servidor MCP que consulta a documentação oficial e atualizada de frameworks e injeta os trechos relevantes no contexto da IA. Evita alucinações com APIs defasadas. O agente usa dois MCPs em conjunto: Context7 (docs) + Playwright (validação).

### Fluxo de dados
```
Prompt com instrução → agente ativa Context7 MCP
  → Context7: busca docs Better Auth (Next.js, GitHub provider, SQLite)
  → Context7: retorna trechos de documentação atuais
  → IA gera código baseado nos docs reais (não no que aprendeu no treino)
  → Playwright MCP valida que o servidor responde na porta correta
```

### Arquivos principais
- [lib/auth.ts](exemplo-08-context7/nextjs-better-auth-demo/lib/auth.ts) — configuração Better Auth com SQLite e GitHub OAuth
- [lib/auth-client.ts](exemplo-08-context7/nextjs-better-auth-demo/lib/auth-client.ts) — client para usar no browser (React)
- [app/api/auth/[...all]/route.ts](exemplo-08-context7/nextjs-better-auth-demo/app/api/auth/%5B...all%5D/route.ts) — route handler catch-all
- [app/page.tsx](exemplo-08-context7/nextjs-better-auth-demo/app/page.tsx) — home com estado de sessão
- [app/login/page.tsx](exemplo-08-context7/nextjs-better-auth-demo/app/login/page.tsx) — botão GitHub OAuth
- [prompt.md](exemplo-08-context7/nextjs-better-auth-demo/prompt.md) — prompt estruturado completo (JSON Prompt de 10 blocos)

### Trechos de código importantes
```typescript
// lib/auth.ts — configuração do servidor Better Auth
// new Database("./better-auth.sqlite") = SQLite local, sem Docker
import { betterAuth } from "better-auth";
import Database from "better-sqlite3";

export const auth = betterAuth({
  database: new Database("./better-auth.sqlite"), // arquivo local SQLite
  socialProviders: {
    github: {
      clientId: process.env.GITHUB_CLIENT_ID as string,
      clientSecret: process.env.GITHUB_CLIENT_SECRET as string,
    },
  },
});

// app/api/auth/[...all]/route.ts — catch-all route para o Better Auth
// Intercepta /api/auth/callback/github, /api/auth/sign-in, etc.
import { auth } from "@/lib/auth";
import { toNextJsHandler } from "better-auth/next-js";
export const { GET, POST } = toNextJsHandler(auth);  // uma linha resolve tudo

// app/page.tsx — React client-side com estado de sessão
"use client";
import { authClient } from "@/lib/auth-client";
const { data: session, isPending } = authClient.useSession();
// session.user.email → usuário logado | null → não logado
```

```markdown
// prompt.md — trecho da regra crítica de uso do Context7
// Demonstra prompt engineering: regras que param o agente se a ferramenta falha
"Você TEM acesso a MCPs no VS Code, e DEVE usar o Context7 MCP.
Regra crítica: Se o Context7 MCP não estiver disponível/funcionando,
PARE o processo imediatamente e responda apenas:
'Context7 MCP não disponível. Não posso continuar.'"
```

### Dependências e como rodar
```bash
npm install
npx @better-auth/cli migrate    # cria tabelas no SQLite
npm run dev                     # http://localhost:3000
```
Variáveis de ambiente (arquivo `.env.local`):
```env
GITHUB_CLIENT_ID=seu_id_aqui
GITHUB_CLIENT_SECRET=seu_secret_aqui
BETTER_AUTH_URL=http://localhost:3000
```
Para criar OAuth App no GitHub: `github.com/settings/developers` → New OAuth App → callback: `http://localhost:3000/api/auth/callback/github`

### O que eu precisaria mudar para adaptar a um projeto próprio
- Adicionar outros providers (Google, Discord) ao `socialProviders`
- Trocar SQLite por PostgreSQL (trocar `better-sqlite3` por adaptador `pg` do Better Auth)
- Adicionar proteção de rotas com middleware Next.js verificando `auth.api.getSession()`
- Estender o schema de usuário (nome, foto, permissões) via configuração do Better Auth

### Dúvidas que um dev backend Java/Spring teria

**1. O route handler `[...all]` é como um `@RequestMapping("/**")` do Spring?**
Exato. O `[...all]` no App Router do Next.js captura qualquer subpath de `/api/auth/*`. Better Auth usa esse endpoint para gerenciar callback OAuth, sign-in, sign-out, session — tudo em um único handler.

**2. `authClient.useSession()` é como um `@AuthenticationPrincipal` do Spring Security?**
Analogia válida. `useSession()` é um React hook que consulta a sessão atual via cookie/token de forma reativa (re-renderiza quando a sessão muda). No Spring, você acessa o `SecurityContext` no servidor; aqui você acessa no cliente de forma declarativa.

**3. Por que o Context7 MCP é necessário? A IA não conhece o Better Auth?**
O modelo foi treinado com dados até certa data — APIs de bibliotecas mudam. Sem Context7, a IA pode gerar código com uma versão antiga da API (ex: parâmetros renomeados, métodos removidos). Context7 injeta o trecho de docs da versão atual, eliminando esse risco.

---

## PROJETO: exemplo-09-grafana-mcp

### O que faz
Stack completa de **observabilidade** com OpenTelemetry, Grafana, Prometheus, Loki e Tempo. A aplicação demo (Fastify + Knex + PostgreSQL) simula um **bug de connection leak** que só é detectável via telemetria — demonstrando como o MCP do Grafana permite que a IA analise métricas e correlacione traces.

### Conceito de IA aplicado
**MCP Grafana** — a IA pode consultar dashboards, queries Prometheus e traces do Tempo através do protocolo MCP. O agente consegue identificar anomalias, correlacionar logs + métricas + traces e sugerir mitigações sem o dev precisar navegar no Grafana manualmente.

### Fluxo de dados
```
Requisição HTTP → Fastify (com OpenTelemetry)
  → OTel SDK: cria span (trace), coleta métrica, emite log
  → OTel Collector: recebe via gRPC 4317
     → Tempo (traces): armazena spans com TraceID
     → Prometheus (metrics): pool de conexões, latência, erros
     → Loki (logs): logs estruturados JSON (pino-loki)
  → Grafana: dashboards unificados
     → MCP Grafana: IA consulta dados via protocolo MCP
     → Agente IA detecta: "connection pool esgotado → connection leak"
     → Agente sugere: rota /reset para liberar conexões
```

### Arquivos principais
- [_alumnus/src/monitoring/otel.ts](_alumnus/src/monitoring/otel.ts) — inicialização completa do SDK OpenTelemetry
- [_alumnus/src/scenarios/db-leaky-connections/main.ts](_alumnus/src/scenarios/db-leaky-connections/main.ts) — bug intencional de connection leak
- [_alumnus/src/database/db.ts](_alumnus/src/database/db.ts) — queries Knex com N+1 intencional (selectAllBadQuery)
- [infra/docker-compose-infra.yaml](exemplo-09-grafana-mcp/infra/docker-compose-infra.yaml) — stack: Grafana, Prometheus, Loki, Tempo, OTel Collector, PostgreSQL
- [docs/prompt.md](exemplo-09-grafana-mcp/docs/prompt.md) — prompts para o agente IA analisar via Grafana MCP

### Trechos de código importantes
```typescript
// otel.ts — inicialização do SDK com traces + metrics + logs
// Exporta tudo para o OTel Collector via gRPC
const _sdk = new NodeSDK({
  serviceName,
  traceExporter: new OTLPTraceExporter(),    // → Tempo
  metricReader: new PeriodicExportingMetricReader({
    exporter: new OTLPMetricExporter(),       // → Prometheus via Collector
    exportIntervalMillis: 3000                // exporta a cada 3 segundos
  }),
  logRecordProcessor: new BatchLogRecordProcessor(
    new OTLPLogExporter()                     // → Loki
  ),
  instrumentations: [
    getNodeAutoInstrumentations({ /* ... */ }),
    new FastifyOtelInstrumentation(),         // auto-instrumenta rotas Fastify
    new KnexInstrumentation(),                // auto-instrumenta queries SQL
  ]
});
_sdk.start();

// db-leaky-connections/main.ts — o bug intencional
// Pool de apenas 2 conexões + não libera após uso = timeout na 3a requisição
this.pool = new pg.Pool({
  connectionString: config.DATABASE_URL,
  max: 2,                          // só 2 conexões permitidas
  connectionTimeoutMillis: 1000    // timeout em 1 segundo se pool estiver cheio
});

// BUG: adquire conexão do pool mas NUNCA chama client.release()
const client = await this.pool.connect();
this.leakedConnections.push(client);  // acumula conexões vazadas
// ... usa o client para query ...
// client.release() AUSENTE — conexão não volta para o pool
```

```typescript
// db.ts — N+1 query intencional para demonstrar em traces
// selectAllBadQuery: 1 query para estudantes + N queries para cursos
export async function selectAllBadQuery(db: Knex): Promise<Student[]> {
  const students = await db('students').select('*');  // 1 query
  for (const student of students) {
    // N queries (uma por estudante) — visível como N spans no Tempo!
    const course = await db('courses').select('*').where({ id: student.courseId }).first();
    student.course = course.name;
  }
  return students;
}

// selectAllGoodQuery: 1 único JOIN — 1 span no Tempo
export async function selectAllGoodQuery(db: Knex): Promise<Student[]> {
  return db('students')
    .select('students.id', 'students.name', 'courses.name as course')
    .innerJoin('courses', 'courses.id', 'students.courseId');
}
```

### Dependências e como rodar
```bash
# Iniciar toda a stack de observabilidade
npm run docker:infra:up      # Grafana :3000, Prometheus :9090, etc.

# Iniciar a aplicação demo
npm start                    # Fastify na porta 9000

# Executar testes (detecção do leak)
npm test

# Parar tudo
npm run docker:infra:down
```
Variáveis de ambiente:
```env
PORT=9000
DATABASE_URL=postgresql://alumnus:alumnus_dev_password@localhost:5433/alumnus_app
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
NODE_ENV=production
LOG_LEVEL=info
```

### O que eu precisaria mudar para adaptar a um projeto próprio
- Trocar Fastify por Express/NestJS (há instrumentações OTel para ambos)
- Ajustar `serviceName` no `otel.ts` para o nome do seu serviço
- Adicionar `KnexInstrumentation` se usar Knex, ou `PgInstrumentation` para pg nativo
- Configurar dashboards no Grafana com suas métricas de negócio

### Dúvidas que um dev backend Java/Spring teria

**1. Isso é equivalente ao Spring Actuator + Micrometer + Zipkin do ecossistema Spring?**
Exatamente. `@OpenTelemetry` + NodeSDK = Micrometer. OTel Collector = intermediário. Tempo = Zipkin/Jaeger. Prometheus = Prometheus (igual). Loki = sem equivalente direto (Splunk/ELK no Java). Grafana = Grafana (igual nos dois mundos).

**2. `@fastify/otel` auto-instrumenta todas as rotas? Preciso anotar cada endpoint?**
Não precisa anotar nada. A auto-instrumentação intercepta automaticamente todas as requisições HTTP (como um Spring AOP ao redor de todos os controllers). O trace inclui método, path, status code, duração — sem código extra nas rotas.

**3. Como correlacionar logs com traces? O `traceId` aparece nos logs automaticamente?**
Com `pino-loki` e a instrumentação OTel, o `traceId` e `spanId` são injetados automaticamente nos logs estruturados. No Grafana, você clica em um trace no Tempo e vê os logs correspondentes no Loki — correlação sem configuração manual.

---

## PROJETO: exemplo-10-ollama

### O que faz
Script bash que demonstra como chamar **modelos de linguagem locais via Ollama**, comparando dois modelos: `llama2-uncensored:7b` (sem filtros de segurança) e `gpt-oss:20b` (com guardrails). Usa a API HTTP do Ollama, compatível com o padrão OpenAI.

### Conceito de IA aplicado
**LLMs locais e quantização.** Ollama gerencia o download, armazenamento e execução de modelos open-source localmente. Demonstra a API REST compatível com OpenAI, streaming de respostas e diferença de comportamento entre modelos com e sem censura.

### Fluxo de dados
```
ollama serve → servidor HTTP local porta 11434
  → ollama pull llama2-uncensored:7b  ← baixa modelo (quantizado)
  → POST /v1/chat/completions         ← API formato OpenAI
     → { model, messages, temperature, max_tokens }
     → resposta JSON com choices[0].message.content
  → POST /api/generate               ← API nativa Ollama
     → { model, prompt, stream }
     → resposta com "response" + "thinking" (para modelos com CoT)
```

### Arquivos principais
- [request.sh](exemplo-10-ollama/request.sh) — script com 3 chamadas: list, chat/completions (llama2), generate (gpt-oss streaming)

### Trechos de código importantes
```bash
# 1. Listar modelos instalados
ollama list

# 2. Baixar modelos
ollama pull llama2-uncensored:7b   # modelo sem filtros de segurança
ollama pull gpt-oss:20b            # modelo com guardrails

# 3. API compatível com OpenAI (/v1/chat/completions)
# Permite trocar de provedor apenas mudando a base URL
curl --silent -X POST http://localhost:11434/v1/chat/completions \
-H "Content-Type: application/json" \
-d '{
  "model": "llama2-uncensored:7b",
  "messages": [{"role":"user","content":"how to create an aim bot on cs 1.6"}]
}' | jq
# Resultado: instruções detalhadas (sem censura)

# 4. API nativa Ollama (/api/generate) com streaming e chain-of-thought
curl --silent -X POST http://localhost:11434/api/generate \
-d '{
  "model": "gpt-oss:20b",
  "prompt": "how to create an aim bot on cs 1.6",
  "stream": false
}' | jq "{response: .response, thinking: .thinking}"
# Resultado: recusa educada + "thinking" mostra o raciocínio interno
```

### Dependências e como rodar
```bash
# Instalar Ollama: https://ollama.com/download
ollama serve                         # inicia servidor local (porta 11434)
ollama pull llama2-uncensored:7b    # ~4GB para 7B quantizado
ollama pull gpt-oss:20b             # ~12GB para 20B

source .env                          # se tiver variáveis
bash request.sh                      # executa as 3 chamadas
```
Requisitos: mínimo 8GB RAM para 7B, 16GB para 20B.
Variáveis de ambiente: **nenhuma** (Ollama roda sem autenticação por padrão).

### O que eu precisaria mudar para adaptar a um projeto próprio
- Trocar `http://localhost:11434` pela URL de um servidor Ollama remoto para compartilhar o modelo
- Mudar o modelo para `llama3.1:8b` ou `mistral:7b` dependendo da tarefa
- Para produção com múltiplos usuários: usar vLLM ou LLaMA.cpp com servidor HTTP ao invés do Ollama (que é single-threaded)
- Integrar com OpenRouter como fallback: se Ollama falhar, roteia para API em nuvem

### Dúvidas que um dev backend Java/Spring teria

**1. A API `/v1/chat/completions` é exatamente igual à da OpenAI? Posso trocar sem mudar código?**
Sim, é drop-in replacement. Mude apenas `baseURL` de `https://api.openai.com` para `http://localhost:11434`. O formato de request/response é idêntico. É por isso que o OpenRouter também usa esse padrão.

**2. Ollama pode rodar como serviço em produção num servidor Linux?**
Sim, mas não é recomendado para alta concorrência — processa uma requisição por vez. Para produção com múltiplos usuários simultâneos, use vLLM (Python) ou Ollama com `--num-parallel` limitado. Para servidores com GPU Nvidia, use `CUDA_VISIBLE_DEVICES` para controlar qual GPU.

**3. Quantização reduz qualidade? Vale a pena usar Q4 vs Q8?**
Q4 (4 bits): 50-60% menor que full precision, perda mínima para tarefas de texto. Q8 (8 bits): 25% menor, qualidade quase idêntica ao original. Regra prática: use Q4F16 para memória limitada, Q8 se tiver RAM suficiente. Para código ou matemática: prefira Q8 (mais preciso).

---

## PROJETO: exemplo-11-openrouter

### O que faz
Script bash minimalista que demonstra como fazer uma chamada à **API do OpenRouter** usando cURL, acessando o modelo Gemma 3 27B gratuitamente. Mostra como configurar headers de autenticação e parâmetros de geração.

### Conceito de IA aplicado
**OpenRouter** — API unificada que roteia requisições para múltiplos provedores de LLM (Google, Meta, Mistral, etc.) com billing centralizado e fallback automático. Usa o mesmo formato da API OpenAI, permitindo troca de modelo por configuração.

### Fluxo de dados
```
source .env                         ← carrega OPENROUTER_API_KEY
  → POST https://openrouter.ai/api/v1/chat/completions
     Headers: Authorization, HTTP-Referer, X-Title
     Body: { model, messages, temperature, max_tokens }
  → OpenRouter roteia para Google Gemma 3 27B
  → resposta JSON → jq (formatação)
```

### Arquivos principais
- [request.sh](exemplo-11-openrouter/request.sh) — único arquivo: cURL com headers OpenRouter + jq

### Trechos de código importantes
```bash
source .env  # carrega variáveis: OPENROUTER_API_KEY

API_URL="https://openrouter.ai/api/v1/chat/completions"
NLP_MODEL="google/gemma-3-27b-it:free"  # modelo gratuito

curl --silent -X POST "$API_URL" \
-H "Content-Type: application/json" \           # tipo de conteúdo
-H "Authorization: Bearer $OPENROUTER_API_KEY" \# autenticação
-H "HTTP-Referer: http://localhost:3000" \      # exigido pelo OpenRouter
-H "X-Title: My Example" \                      # identificação da app
-d '{
  "model": "'"$NLP_MODEL"'",
  "messages": [{"role":"user","content":"Me conte uma curiosidade sobre LLMs"}],
  "temperature": 0.3,    # baixo = mais determinístico
  "max_tokens": 1000
}' | jq   # formata JSON no terminal
```

### Dependências e como rodar
```bash
# Obter API key gratuita: https://openrouter.ai
# Criar arquivo .env:
echo "OPENROUTER_API_KEY=sk-or-v1-..." > .env

source .env
bash request.sh
```
Variáveis de ambiente: `OPENROUTER_API_KEY` (obrigatório)

### O que eu precisaria mudar para adaptar a um projeto próprio
- Trocar `NLP_MODEL` por qualquer modelo disponível no OpenRouter (llama, mistral, claude, gpt-4o)
- Para Node.js: substituir cURL por `fetch()` com os mesmos headers
- Para Java/Spring: usar `RestTemplate` ou `WebClient` — mesmo formato de request
- Adicionar `"stream": true` para streaming de tokens

### Dúvidas que um dev backend Java/Spring teria

**1. Por que os headers `HTTP-Referer` e `X-Title` são obrigatórios?**
O OpenRouter usa para identificar de qual aplicação vem a requisição — permite rastrear uso, aplicar rate limits por app e exibir estatísticas no dashboard. Sem eles, a requisição pode ser rejeitada dependendo da política do modelo.

**2. O modelo `:free` tem limitações? Precisa pagar para produção?**
Modelos `:free` têm rate limits mais baixos (ex: 20 req/min) e podem ter menor disponibilidade. Para produção, use modelos pagos com SLA definido. O OpenRouter tem billing unificado — um cartão paga múltiplos provedores.

**3. Como integraria isso num microserviço Spring Boot?**
```java
// Spring Boot com RestTemplate (exemplo conceitual)
HttpHeaders headers = new HttpHeaders();
headers.set("Authorization", "Bearer " + apiKey);
headers.set("HTTP-Referer", "https://meuapp.com");
headers.set("X-Title", "Meu App");
HttpEntity<Map> request = new HttpEntity<>(body, headers);
ResponseEntity<String> response = restTemplate.postForEntity(
    "https://openrouter.ai/api/v1/chat/completions", request, String.class);
```
Para streaming, use `WebClient` com `Flux<ServerSentEvent>`.

---

## PROJETO: exemplo-12-embeddings-neo4j-template

### O que faz
Template inicial para implementar busca semântica com embeddings. Contém a configuração completa (`config.ts`), utilitários de exibição (`util.ts`) e infraestrutura Docker (Neo4j), mas o `index.ts` e o `DocumentProcessor` estão para o aluno implementar.

> **Comparação com exemplo-12**: este template é o ponto de partida; o exemplo-12 (sem `-template`) tem a implementação completa com `DocumentProcessor` e pipeline de embeddings funcionando.

### Conceito de IA aplicado
**Embeddings e Vector Store.** A configuração já define: modelo de embedding (HuggingFace), conexão Neo4j, chunking de documentos e busca por similaridade. O aluno implementa o pipeline que conecta esses componentes.

### Fluxo de dados (a implementar)
```
PDF (tensores.pdf)
  → [IMPLEMENTAR] DocumentProcessor.loadAndSplit()
  → chunks de texto
  → [IMPLEMENTAR] HuggingFaceTransformersEmbeddings
  → vetores numéricos
  → [IMPLEMENTAR] Neo4jVectorStore.addDocuments()
  → [IMPLEMENTAR] similaritySearch(question, topK)
  → resultados exibidos por util.ts (já pronto)
```

### Arquivos principais
- [src/config.ts](exemplo-12-embeddings-neo4j-template/src/config.ts) — todas as configs em um objeto `CONFIG` frozen
- [src/util.ts](exemplo-12-embeddings-neo4j-template/src/util.ts) — `displayResults()` pronta para uso
- [src/index.ts](exemplo-12-embeddings-neo4j-template/src/index.ts) — **vazio** — a implementar
- [docker-compose.yml](exemplo-12-embeddings-neo4j-template/docker-compose.yml) — Neo4j 5.14 com plugin APOC

### Trechos de código importantes
```typescript
// src/config.ts — configuração centralizada (já pronta no template)
export const CONFIG = Object.freeze({
  neo4j: {
    url: process.env.NEO4J_URI!,
    username: process.env.NEO4J_USER!,
    password: process.env.NEO4J_PASSWORD!,
    indexName: "tensors_index",
    searchType: "vector" as const,   // busca por vetor (não texto)
    textNodeProperties: ["text"],    // propriedade do nó que contém o texto
    nodeLabel: "Chunk",              // label dos nós no grafo
  },
  embedding: {
    modelName: process.env.EMBEDDING_MODEL!,  // ex: Xenova/all-MiniLM-L6-v2
    pretrainedOptions: {
      dtype: "fp32" as DataType,    // qualidade vs velocidade: fp32 > fp16 > q8 > q4
    },
  },
  textSplitter: {
    chunkSize: 1000,     // tamanho máximo de cada chunk em caracteres
    chunkOverlap: 200,   // overlap entre chunks para não perder contexto
  },
  similarity: {
    topK: 3,  // retorna os 3 chunks mais similares
  },
});
```

### Dependências e como rodar
```bash
cp .env.example .env    # configurar variáveis
npm ci
npm run infra:up        # sobe Neo4j via Docker
# implementar index.ts e documentProcessor.ts
npm start
```
Variáveis de ambiente:
```env
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password
EMBEDDING_MODEL=Xenova/all-MiniLM-L6-v2
```

### O que eu precisaria mudar para adaptar a um projeto próprio
- Apontar `CONFIG.pdf.path` para seus documentos
- Ajustar `chunkSize` e `chunkOverlap` para o tipo do seu conteúdo (docs técnicos: 500-800, narrativos: 1000-1500)
- Trocar `EMBEDDING_MODEL` por outro modelo da HuggingFace adequado ao idioma (ex: `neuralmind/bert-base-portuguese-cased` para PT-BR especializado)

### Dúvidas que um dev backend Java/Spring teria

**1. `Object.freeze(CONFIG)` é como `@Value` imutável no Spring?**
Sim. `Object.freeze` impede modificações no objeto em runtime — equivalente a propriedades `final` injetadas via `@Value` ou `@ConfigurationProperties`. Garante que a config não é mutada acidentalmente.

**2. `chunkOverlap: 200` — por que o overlap é necessário?**
Ao dividir o documento em pedaços de 1000 chars, o contexto semântico pode ser cortado no meio de uma frase. Com overlap de 200 chars, o início de cada chunk repete o final do anterior — preservando contexto. Equivale a ler um livro com 200 caracteres de "recap" no início de cada capítulo.

**3. Neo4j como vector database é diferente de Neo4j como banco de grafos?**
É o mesmo banco. O Neo4j 5.x adicionou suporte nativo a índices vetoriais. Você pode ter nós com embeddings e fazer busca por similaridade (`db.index.vector.queryNodes`) E ao mesmo tempo fazer queries de grafo (relações entre chunks, metadados). É um banco híbrido grafo+vetorial.

---

## PROJETO: exemplo-12-embeddings-neo4j

### O que faz
Implementação completa da busca semântica com embeddings: carrega um PDF sobre TensorFlow.js, divide em chunks, gera embeddings localmente com HuggingFace Transformers, armazena no Neo4j e executa 6 buscas por similaridade semântica — tudo sem API externa, sem custo.

### Conceito de IA aplicado
**Embeddings + Vector Store + Busca por Similaridade Semântica.** O modelo converte texto em vetores numéricos; textos com significado similar ficam próximos no espaço vetorial. A busca por similaridade encontra os chunks relevantes mesmo que as palavras exatas sejam diferentes da pergunta.

### Fluxo de dados
```
tensores.pdf
  → PDFLoader → documentos brutos por página
  → RecursiveCharacterTextSplitter → N chunks (1000 chars, overlap 200)
  → HuggingFaceTransformersEmbeddings.embedDocuments()
     → modelo local Xenova/all-MiniLM-L6-v2 (sem API)
     → vetor de 384 dimensões por chunk
  → Neo4jVectorStore.addDocuments() → armazena chunk + embedding como nó
  → similaritySearch(question, topK=3)
     → embedding da pergunta → busca por coseno no índice vetorial
     → retorna top 3 chunks mais próximos semanticamente
  → displayResults() → exibe trechos relevantes no terminal
```

### Arquivos principais
- [src/index.ts](exemplo-12-embeddings-neo4j/src/index.ts) — pipeline completo com clearAll + addDocuments + similaritySearch
- [src/documentProcessor.ts](exemplo-12-embeddings-neo4j/src/documentProcessor.ts) — `DocumentProcessor` com PDFLoader + RecursiveCharacterTextSplitter
- [src/config.ts](exemplo-12-embeddings-neo4j/src/config.ts) — CONFIG centralizado (idêntico ao template)
- [src/util.ts](exemplo-12-embeddings-neo4j/src/util.ts) — `displayResults()` com formatação

### Trechos de código importantes
```typescript
// documentProcessor.ts — carrega PDF e divide em chunks
export class DocumentProcessor {
  async loadAndSplit() {
    const loader = new PDFLoader(this.pdfPath);
    const rawDocuments = await loader.load();  // uma página = um documento
    console.log(`Loaded ${rawDocuments.length} pages from PDF`);

    // RecursiveCharacterTextSplitter: tenta quebrar em parágrafos, depois frases
    const splitter = new RecursiveCharacterTextSplitter({
      chunkSize: 1000,     // max 1000 chars por chunk
      chunkOverlap: 200    // 200 chars de sobreposição entre chunks adjacentes
    });

    const documents = await splitter.splitDocuments(rawDocuments);
    console.log(`Split into ${documents.length} chunks`);
    return documents;
  }
}

// index.ts — pipeline principal
const embeddings = new HuggingFaceTransformersEmbeddings({
  model: CONFIG.embedding.modelName,  // "Xenova/all-MiniLM-L6-v2"
  // Roda localmente — baixa o modelo na 1a execução (~25MB)
});

// Conecta ao Neo4j usando índice vetorial existente (ou cria)
const vectorStore = await Neo4jVectorStore.fromExistingGraph(embeddings, CONFIG.neo4j);

// Limpa dados antigos antes de reinserir
await vectorStore.query(`MATCH (n:\`Chunk\`) DETACH DELETE n`);

// Adiciona documentos um a um (para progresso visível)
for (const [index, doc] of documents.entries()) {
  await vectorStore.addDocuments([doc]);  // gera embedding e armazena no Neo4j
}

// Busca semântica — funciona com paráfrase, sinônimos, conceitos relacionados
const results = await vectorStore.similaritySearch(
  "O que são tensores e como são representados em JavaScript?",
  3  // topK = retorna 3 chunks mais próximos
);
```

### Dependências e como rodar
```bash
npm ci
npm run infra:up           # sobe Neo4j Docker (aguarda health check)

# Criar .env com:
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password
EMBEDDING_MODEL=Xenova/all-MiniLM-L6-v2

npm start                  # baixa modelo HF na 1a execução (~30s)
npm run infra:down         # para o Neo4j
```

### O que eu precisaria mudar para adaptar a um projeto próprio
- Trocar `tensores.pdf` por seus documentos (manuais, contratos, base de conhecimento)
- Substituir `PDFLoader` por outros loaders LangChain: `DirectoryLoader`, `CSVLoader`, `JSONLoader`
- Ajustar `chunkSize` conforme o tipo de documento
- Para produção: extrair o passo de indexação (addDocuments) para um job batch separado da busca

### Dúvidas que um dev backend Java/Spring teria

**1. O modelo Xenova/all-MiniLM-L6-v2 é bom para português?**
Foi treinado principalmente em inglês e multilíngue, mas performa razoavelmente em PT-BR. Para qualidade máxima em português, use `neuralmind/bert-base-portuguese-cased` ou `sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2`.

**2. `--experimental-strip-types` do Node.js é seguro para produção?**
Para projetos internos com controle do ambiente (Node.js 22.x fixo), sim. Para produção pública, prefira compilar com `tsc` ou usar `tsx`. O strip-types apenas remove as anotações TypeScript sem verificar tipos — é mais rápido mas sem type-check.

**3. Armazenar embeddings no Neo4j vs num banco vetorial dedicado (Pinecone) — quando usar cada um?**
Neo4j: ideal quando você já tem dados no grafo (relações entre entidades) e quer adicionar busca semântica. Pinecone/Qdrant: otimizados exclusivamente para busca vetorial, mais rápidos em escala. Para começar: Neo4j. Para 10M+ vetores em produção: banco vetorial dedicado.

---

## PROJETO: exemplo-13-embeddings-neo4j-rag

### O que faz
Pipeline **RAG completo** (Retrieval-Augmented Generation): carrega PDF, gera embeddings, armazena no Neo4j, e para cada pergunta busca os chunks relevantes e envia para uma LLM (via OpenRouter) gerar uma resposta fundamentada em texto real — salva as respostas em arquivos Markdown.

> **Evolução em relação ao exemplo-12**: o exemplo-12 só faz busca por similaridade (retorna chunks brutos). Este exemplo adiciona a geração de resposta por LLM com RAG completo, prompts estruturados em JSON e persistência das respostas.

### Conceito de IA aplicado
**RAG (Retrieval-Augmented Generation).** A LLM não "inventa" a resposta — ela fundamenta a geração nos chunks reais recuperados do Neo4j. O pipeline usa `RunnableSequence` do LangChain para encadear: busca vetorial → filtragem por score → geração com LLM → persistência.

### Fluxo de dados
```
tensores.pdf → chunks → embeddings → Neo4j (igual ao exemplo-12)

Para cada pergunta:
  → AI.answerQuestion(question)
     → RunnableSequence([
         retrieveVectorSearchResults,  // busca top-3 no Neo4j
         generateNLPResponse           // gera resposta com LLM
       ])
     → similaritySearchWithScore(question, topK=3)
        → filtra chunks com score > 0.5
        → concatena contexto dos chunks relevantes
     → ChatOpenAI(OpenRouter/Gemma 3 27B)
        → prompt estruturado: role + task + tone + context + question
        → resposta em PT-BR
     → writeFile(./respostas/resposta-N-timestamp.md)
```

### Arquivos principais
- [src/ai.ts](exemplo-13-embeddings-neo4j-rag/src/ai.ts) — classe `AI` com `RunnableSequence`, busca + geração
- [src/index.ts](exemplo-13-embeddings-neo4j-rag/src/index.ts) — pipeline completo integrando `AI` + Neo4j + persistência
- [src/config.ts](exemplo-13-embeddings-neo4j-rag/src/config.ts) — config + leitura dos arquivos de prompt
- [src/documentProcessor.ts](exemplo-13-embeddings-neo4j-rag/src/documentProcessor.ts) — idêntico ao exemplo-12
- [prompts/answerPrompt.json](exemplo-13-embeddings-neo4j-rag/prompts/answerPrompt.json) — prompt estruturado em JSON (role, task, instructions, constraints)
- [prompts/template.txt](exemplo-13-embeddings-neo4j-rag/prompts/template.txt) — template com placeholders para o ChatPromptTemplate

### Trechos de código importantes
```typescript
// ai.ts — RunnableSequence encadeia 2 funções como pipeline
// Equivalente a uma chain de Filters no Java ou ao padrão Pipeline
const chain = RunnableSequence.from([
  this.retrieveVectorSearchResults.bind(this),  // passo 1: busca
  this.generateNLPResponse.bind(this)           // passo 2: gera
]);
const result = await chain.invoke({ question });

// Passo 1: busca vetorial com score de confiança
async retrieveVectorSearchResults(input: ChainState): Promise<ChainState> {
  const vectorResults = await this.params.vectorStore
    .similaritySearchWithScore(input.question, this.params.topK);

  const topScore = vectorResults[0]![1];  // score do melhor resultado (0-1)

  // Filtra apenas resultados com alta confiança (score > 50%)
  const contexts = vectorResults
    .filter(([, score]) => score > 0.5)   // descarta chunks irrelevantes
    .map(([doc]) => doc.pageContent)
    .join("\n\n---\n\n");                  // separa chunks com marcador

  return { ...input, context: contexts, topScore };
}

// Passo 2: gera resposta com LLM usando contexto recuperado
async generateNLPResponse(input: ChainState): Promise<ChainState> {
  if (input.error) return input;  // curto-circuito se não há contexto válido

  // Template com placeholders — preenchido com dados do answerPrompt.json
  const responsePrompt = ChatPromptTemplate.fromTemplate(this.params.templateText);
  const chain = responsePrompt
    .pipe(this.params.nlpModel)      // ChatOpenAI apontando para OpenRouter
    .pipe(new StringOutputParser()); // extrai string da resposta

  const rawResponse = await chain.invoke({
    role: this.params.promptConfig.role,
    task: this.params.promptConfig.task,
    context: input.context,           // chunks recuperados do Neo4j
    question: input.question,         // pergunta original do usuário
    // ... demais campos do prompt
  });

  return { ...input, answer: rawResponse };
}
```

```json
// prompts/answerPrompt.json — prompt estruturado (JSON Prompt)
// Separa cada componente do prompt em campo nomeado — reduz ambiguidade
{
  "task": "Responder perguntas sobre TensorFlow.js e machine learning",
  "role": "Você é um assistente especializado em TensorFlow.js",
  "instructions": [
    "Use APENAS as informações do contexto fornecido para responder",
    "Se o contexto não contiver informação suficiente, diga que não encontrou",
    "Responda em português de forma natural e conversacional"
  ],
  "constraints": {
    "language": "pt-BR",
    "tone": "educacional e amigável",
    "max_length": 500,
    "format": "texto natural com exemplos"
  }
}
```

### Dependências e como rodar
```bash
npm ci
npm run infra:up

# .env:
NEO4J_URI=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password
EMBEDDING_MODEL=Xenova/all-MiniLM-L6-v2
NLP_MODEL=google/gemma-3-27b-it:free
OPENROUTER_API_KEY=sk-or-v1-...
OPENROUTER_SITE_URL=http://localhost:3000
OPENROUTER_SITE_NAME=My Example

npm start    # gera embeddings + responde 5 perguntas + salva em ./respostas/
```

### O que eu precisaria mudar para adaptar a um projeto próprio
- Trocar o PDF por seus documentos de domínio (manuais, políticas, base de conhecimento)
- Editar `prompts/answerPrompt.json` com o `role`, `task` e `instructions` do seu domínio
- Ajustar o `threshold` de 0.5 para o seu caso (0.7+ para respostas mais precisas, 0.3 para mais abrangentes)
- Adicionar os metadados relevantes nos chunks (autor, data, seção) para citar fontes na resposta

### Dúvidas que um dev backend Java/Spring teria

**1. `RunnableSequence` do LangChain é como um `@Bean Chain` do Spring Batch?**
Boa analogia. É similar ao `Step` encadeado no Spring Batch — cada etapa recebe o output da anterior. A diferença é que `RunnableSequence` é assíncrono e tipado com generics. Equivale também ao padrão Chain of Responsibility, onde cada handler pode modificar e passar adiante o estado.

**2. Por que salvar as respostas em arquivos .md em vez de exibir no terminal?**
Para auditoria e revisão posterior. Em sistemas RAG de produção, você quer logar: a pergunta, o contexto recuperado (chunks + scores), a resposta gerada e o timestamp. Isso permite avaliar qualidade, detectar alucinações e melhorar o sistema iterativamente.

**3. O filtro `score > 0.5` é suficiente? Como calibrar esse threshold?**
Depende do modelo de embedding e do domínio. Para começar: 0.5 é conservador. Rode algumas consultas e inspecione os scores — se os resultados relevantes têm score 0.3-0.4, baixe o threshold. Se há muito "ruído" (resultados irrelevantes), suba para 0.7. Avalie empiricamente com o seu corpus.

---

## ÍNDICE GERAL

| Projeto | Conceito Principal | Tecnologias |
|---------|-------------------|-------------|
| exemplo-00-template | Rede Neural (template) | TensorFlow.js-node, One-Hot Encoding |
| exemplo-00-z | Rede Neural Multiclasse | TensorFlow.js-node, ReLU, Softmax, Adam |
| exemplo-01-ecommerce-recomendations-template | Sistema de Recomendação (template MVC) | TensorFlow.js (CDN), Vanilla JS, browser-sync |
| exemplo-01-ecommerce-recomendations-z | Sistema de Recomendação Completo | TensorFlow.js, Web Workers, tfvis, MVC |
| exemplo-02-vencendo-qualquer-jogo | YOLO — Detecção de Objetos | PixiJS, TensorFlow.js, YOLO, Webpack, Babel |
| exemplo-03-webai01 | LLM Nativa no Browser (básico) | Chrome Web AI API, LanguageModel, Streaming |
| exemplo-04-webai02-temperature-and-topK | Parâmetros de Sampling de LLM | Chrome Web AI API, AbortController, http-server |
| exemplo-05-webai03-multimodal | LLM Multimodal + Tradução | Chrome Web AI API, Translator API, MVC |
| exemplo-06-playwright-testes | MCP Playwright (geração de testes) | Playwright MCP, @playwright/test, GitHub Actions |
| exemplo-07-playwright-navegacao | MCP Playwright (automação web) | Playwright MCP, automação de formulários |
| exemplo-08-context7 | MCP Context7 + OAuth | Next.js 16, Better Auth, SQLite, Tailwind, MCP |
| exemplo-09-grafana-mcp | OpenTelemetry + Observabilidade | Fastify, OTel SDK, Grafana, Prometheus, Loki, Tempo |
| exemplo-10-ollama | LLMs Locais | Ollama, llama2, gpt-oss, API REST compatível OpenAI |
| exemplo-11-openrouter | API Unificada de LLMs | OpenRouter, cURL, Gemma 3 27B, bash |
| exemplo-12-embeddings-neo4j-template | Embeddings + Vector Store (template) | LangChain, HuggingFace Transformers, Neo4j, Docker |
| exemplo-12-embeddings-neo4j | Embeddings + Busca Semântica | LangChain, HuggingFace, Neo4j, pdf-parse, TypeScript |
| exemplo-13-embeddings-neo4j-rag | RAG Completo | LangChain, HuggingFace, Neo4j, OpenRouter, ChatOpenAI |

---

*Relatório gerado em 05/03/2026 — todos os arquivos dos projetos foram lidos e analisados.*
