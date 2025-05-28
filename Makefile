.PHONY: build push login

IMAGE_NAME = conversor-pdf-
REPO = quay.io/uemcpa

build:
	@read -p "Enter version tag for the image: " VERSION; \
	echo "Building image $(IMAGE_NAME)backend:$$VERSION"; \
	podman build -t $(IMAGE_NAME)backend:$$VERSION -t $(IMAGE_NAME)backend:latest ./backend; \
	echo $$VERSION > .version; \
	echo "Building image $(IMAGE_NAME)frontend:$$VERSION"; \
	podman build -t $(IMAGE_NAME)frontend:$$VERSION -t $(IMAGE_NAME)frontend:latest ./frontend; 

login:
	@echo "Fazendo login no repositório $(REPO)..."; \
	podman login quay.io  --authfile ./auth.json; 

push: login build
	@VERSION=$$(cat .version); \
	echo "Pushing image $(REPO)/$(IMAGE_NAME)backend:$$VERSION"; \
	podman push $(IMAGE_NAME)backend:$$VERSION $(REPO)/$(IMAGE_NAME)backend:$$VERSION; \
	echo "Pushing image $(REPO)/$(IMAGE_NAME)frontend:latest"; \
	podman push $(IMAGE_NAME)frontend:latest $(REPO)/$(IMAGE_NAME)frontend:latest