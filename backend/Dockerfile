# FROM docker.io/pandoc/core:latest-ubuntu
# FROM docker.io/pandoc/extra:3.5.0.0-ubuntu
FROM docker.io/pandoc/extra:latest-ubuntu
LABEL mainatainer="Dioclecio Camelo <dmcamelo@uem.br>"
LABEL br.uem.cpa.relatorios.backend.author="Dioclecio Camelo"
# LABEL br.uem.cpa.relatorios.backend.version="v011"
RUN apt update && apt install -y texlive texlive-latex-extra texlive-fonts-extra texlive-fonts-recommended
WORKDIR /app
COPY ./main /app/main
COPY ./bg /app/bg
EXPOSE 5000
VOLUME /data
ENTRYPOINT [ "/app/main" ]