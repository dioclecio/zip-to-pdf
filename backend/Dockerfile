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
RUN go build main.go
# Montar a imagem que vai para o repositório
FROM docker.io/pandoc/extra:latest-ubuntu
LABEL mainatainer="Dioclecio Camelo <dmcamelo@uem.br>"
LABEL br.uem.cpa.relatorios.backend.author="Dioclecio Camelo"
RUN mkdir /root/.fonts/ && apt update && apt install -y texlive texlive-latex-extra texlive-fonts-extra texlive-fonts-recommended
WORKDIR /app
# COPY ./main /app/main
COPY ./bg /app/bg
COPY --from=unzipper /opt/.fonts /root/.fonts
COPY --from=compiler /src/main .
EXPOSE 5000
VOLUME /data
ENTRYPOINT [ "/app/main" ]