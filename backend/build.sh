#!/bin/bash
version="v$(date +'%Y%m%d')"
# Faz o login para enviar aos repositórios públicos
podman login quay.io --authfile ../auth.json

# Limpa os repositórios locais
podman build . -t conversor-pdf-backend:latest -t conversor-pdf-backend:$version 

podman push --authfile ../auth.json conversor-pdf-backend:$version quay.io/uemcpa/conversor-pdf-backend:$version
podman push --authfile ../auth.json conversor-pdf-backend:latest quay.io/uemcpa/conversor-pdf-backend:latest
