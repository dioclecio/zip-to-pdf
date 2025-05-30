# Extrair as fontes
FROM docker.io/alpine AS unzipper
WORKDIR /opt
COPY ./fonts .
RUN apk add unzip ; mkdir .fonts
RUN  for i in *.zip; do bn=$(basename $i .zip); mkdir .fonts/$bn; unzip -j -d .fonts/$bn $i ; done
# Compilar o main.go 
FROM docker.io/golang AS compiler
WORKDIR /src
COPY . .
RUN go mod tidy
RUN go build main.go
# Criar a imagem final
# Montar a imagem que vai para o repositório
FROM docker.io/pandoc/latex:latest-ubuntu
LABEL maintainer="Dioclecio Camelo <dmcamelo@uem.br>"
LABEL br.uem.cpa.relatorios.backend.author="Dioclecio Camelo"
RUN mkdir /root/.fonts/ && apt update && apt install -y texlive texlive-latex-extra pdftk
RUN tlmgr install soul adjustbox babel-german background bidi collectbox csquotes everypage filehook footmisc footnotebackref framed fvextra letltxmacro ly1 mdframed mweights needspace pagecolor sourcecodepro sourcesanspro titling ucharcat ulem unicode-math upquote xecjk xurl zref
WORKDIR /app
COPY ./bg /app/bg
COPY --from=unzipper /opt/.fonts /root/.fonts
COPY --from=compiler /src/main .
EXPOSE 5000
VOLUME /data
ENTRYPOINT [ "/app/main" ]