#!/bin/bash
version="v$(date +'%Y%m%d')"
# Limpa os repositórios locais
podman rmi conversor-pdf-frontend:$version
podman rmi conversor-pdf-frontend:latest
podman build . -t conversor-pdf-frontend:$version -t conversor-pdf-frontend:latest

# Faz o login para enviar aos repositórios públicos
podman login quay.io --authfile ../auth.json

podman push conversor-pdf-frontend:$version quay.io/uemcpa/conversor-pdf-frontend:$version
podman push conversor-pdf-frontend:latest quay.io/uemcpa/conversor-pdf-frontend:latest