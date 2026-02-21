# Demo Next.js + Better Auth + GitHub OAuth

Demo simples de autenticação com GitHub usando Next.js (App Router), Better Auth e SQLite.

## 🚀 Funcionalidades

- Login/Signup via GitHub OAuth
- Sessão persistente no banco SQLite local
- Interface limpa com Tailwind CSS
- TypeScript

## 📋 Pré-requisitos

- Node.js 18+ e npm
- Conta GitHub para criar OAuth App

## 🔧 Configuração

### 1. Criar GitHub OAuth App

1. Acesse [GitHub Developer Settings](https://github.com/settings/developers)
2. Clique em "New OAuth App"
3. Preencha:
   - **Application name**: `Next.js Better Auth Demo`
   - **Homepage URL**: `http://localhost:3000`
   - **Authorization callback URL**: `http://localhost:3000/api/auth/callback/github`
4. Copie o `Client ID` e gere um `Client Secret`

### 2. Configurar variáveis de ambiente

Edite o arquivo `.env.local` e adicione suas credenciais:

```bash
GITHUB_CLIENT_ID=seu_client_id_aqui
GITHUB_CLIENT_SECRET=seu_client_secret_aqui
NEXT_PUBLIC_BASE_URL=http://localhost:3000
```

### 3. Instalar dependências

```bash
npm install
```

### 4. Gerar schema do banco de dados

```bash
npx @better-auth/cli migrate
```

Este comando cria o arquivo `better-auth.sqlite` com as tabelas necessárias.

### 5. Iniciar o servidor

```bash
npm run dev
```

Acesse [http://localhost:3000](http://localhost:3000)

## 📁 Estrutura do projeto

```
.
├── app/
│   ├── api/auth/[...all]/route.ts  # Endpoint Better Auth
│   ├── layout.tsx                   # Layout raiz
│   ├── page.tsx                     # Página principal
│   └── globals.css                  # Estilos globais
├── lib/
│   ├── auth.ts                      # Configuração Better Auth (servidor)
│   └── auth-client.ts               # Cliente Better Auth (navegador)
├── .env.local                       # Variáveis de ambiente (não commitado)
├── .env.example                     # Exemplo de variáveis de ambiente
├── package.json
├── tsconfig.json
├── tailwind.config.js
└── next.config.js
```

## 🎯 Como funciona

1. **Página inicial**: Mostra status de login e botão "Entrar com GitHub"
2. **Login**: Clique no botão para iniciar OAuth com GitHub
3. **Callback**: GitHub redireciona para `/api/auth/callback/github`
4. **Sessão**: Better Auth cria sessão no SQLite e armazena cookie
5. **Home**: Exibe dados do usuário logado
6. **Logout**: Botão "Sair" encerra a sessão

## 🛠️ Tecnologias

- **Next.js 15**: Framework React
- **Better Auth**: Biblioteca de autenticação
- **better-sqlite3**: Driver SQLite para Node.js
- **Tailwind CSS**: Estilização
- **TypeScript**: Tipagem estática

## 📝 Notas

- O banco SQLite é criado localmente em `better-auth.sqlite`
- As sessões são persistidas entre recarregamentos
- Em produção, configure `NEXT_PUBLIC_BASE_URL` com sua URL real
