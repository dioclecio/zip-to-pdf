#!/bin/bash
version="v$(date +'%Y%m%d')"

# Faz o login para enviar aos repositórios públicos
podman login quay.io --authfile ../auth.json

# Limpa os repositórios locais
podman build . -t conversor-pdf-frontend:latest -t conversor-pdf-frontend:$version 

podman push --authfile ../auth.json conversor-pdf-frontend:$version quay.io/uemcpa/conversor-pdf-frontend:$version
podman push --authfile ../auth.json conversor-pdf-frontend:latest quay.io/uemcpa/conversor-pdf-frontend:latest
