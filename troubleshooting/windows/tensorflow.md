# Suporte para TensorFlow.js no Windows

> Caminho feliz:

- Use **WSL**, até a próxima 🐧✨
- [Veja como configurar o **WSL** pela primeira vez para atender os requisitos dos módulos](./wsl.md).

---

> Usando diretamente com **Windows** _(experimental)_:

## Instalação do [**Node.js**](https://nodejs.org/) via [nvm](https://github.com/coreybutler/nvm-windows)

Para usar o `@tensorflow/tfjs-node` sem o **WSL** no **Windows**, é necessário configurar o ambiente antes de continuar.

> [!IMPORTANT]
>
> - Se já tiver o **Node.js** na versão **24**, você pode pular essa etapa.
> - Caso já use o **Node.js** em outra versão (sem **nvm**), desinstale o **Node.js** antes de continuar.

> [!NOTE]
>
> - O **tensorflowjs** não possui suporte para o **Node.js** na versão **22**. Por isso, é recomendado usar a versão **24**.

- Baixe o instalador do **nvm** para **Windows**:
  - 📦 https://github.com/coreybutler/nvm-windows/releases/download/1.2.2/nvm-setup.exe
- Execute o instalador e siga as instruções na tela até concluir a instalação do **nvm**.
- Após a instalação, feche qualquer terminal aberto, abra um novo e execute:
  - ```cmd
    nvm install 24
    nvm use 24
    npm i -g yarn
    ```

---

### Compiladores C++ e [Python](https://www.python.org/)

#### Visual Studio Build Tools for C++

- Baixe o instalador do **Visual Studio Build Tools** para Windows:
  - 📦 https://aka.ms/vs/stable/vs_BuildTools.exe
- Caso já esteja instalado, clique na opção **`Modificar/Modify`**.
- Ative:
  - **Desenvolvimento para desktop com C++**
    - Nos detalhes da instalação, marque:
      - **MSVC v143 - VS 2022 C++ x64/x86 build tools**
  - **Node.js build tools**

#### Python

- Baixe o instalador do **Python** para **Windows**:
  - 📦 https://www.python.org/ftp/python/pymanager/python-manager-25.2.msix
- Após a instalação, irá ser aberto um terminal com algumas etapas de "Sim" (<kbd>**Y**</kbd>) ou "Não" (<kbd>**N**</kbd>):
  - Update setting now? <kbd>**Y**</kbd>
  - Add commands directory to your PATH now? <kbd>**Y**</kbd>
  - Install CPython now? <kbd>**Y**</kbd>
  - View online help? <kbd>**N**</kbd>

---

Para verificar a instalação e as versões:

```cmd
node -v
python --version
yarn -v
```

---

Agora você pode usar o **TensorFlow.js** (`@tensorflow/tfjs-node`) normalmente 🤟

- Em algumas distribuições do **Windows** ou processadores, o `@tensorflow/tfjs-node` não pode não possuir os binários compatíveis. Nesse caso, a alternativa mais fácil é usar o `@tensorflow/tfj` que não depende de binários e funciona com **JavaScript** puro, mas sabendo que a performance é menor e que enquanto o `@tensorflow/tfjs-node` usa **C++**, o `@tensorflow/tfj` é voltado para uso em navegadores, podendo divergir em operaçõwes de ponto flutuante (_floats_).
