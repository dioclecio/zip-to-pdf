# FROM docker.io/pandoc/core:latest-ubuntu
FROM docker.io/pandoc/extra:3.5.0.0-ubuntu
LABEL mainatainer="Dioclecio Camelo <dmcamelo@uem.br>"
LABEL br.uem.cpa.relatorios.backend.author="Dioclecio Camelo"
LABEL br.uem.cpa.relatorios.backend.version="v10"
# CMD [ "apt", "update", "&&", "apt", "install", "pandoc" ]
WORKDIR /app
COPY ./main /app/main
EXPOSE 5000
VOLUME /data
ENTRYPOINT [ "/app/main" ]