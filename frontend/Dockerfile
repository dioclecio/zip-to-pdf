# Usar a imagem base do Node.js
FROM docker.io/node
LABEL mainatainer="Dioclecio Camelo <dmcamelo@uem.br>"
LABEL br.uem.cpa.relatorios.frontend.author="Dioclecio Camelo"
LABEL br.uem.cpa.relatorios.frontend.version="v10"
# Criar diretório de trabalho
WORKDIR /app

# Copiar package.json e package-lock.json
COPY package*.json ./

# Instalar dependências
RUN npm install

# Copiar o restante do código da aplicação
COPY . .

# Expor a porta que a aplicação vai rodar
EXPOSE 3000

# Comando para iniciar a aplicação
ENTRYPOINT ["node", "server.js"]