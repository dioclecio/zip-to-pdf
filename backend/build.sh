#!/bin/bash
version="v$(date +'%Y%m%d')"
# Compila o arquivo
podman run --name compilador \
    --rm -it -v "$PWD":/usr/src/app \
    -w /usr/src/app \
    docker.io/golang go build main.go
    
# Limpa os repositórios locais
podman rmi conversor-pdf-backend:$version
podman rmi conversor-pdf-backend:latest
podman build . -t conversor-pdf-backend:$version -t conversor-pdf-backend:latest

# Faz o login para enviar aos repositórios públicos
podman login quay.io --authfile ../auth.json

podman push conversor-pdf-backend:$version quay.io/uemcpa/conversor-pdf-backend:$version
podman push conversor-pdf-backend:latest quay.io/uemcpa/conversor-pdf-backend:latest