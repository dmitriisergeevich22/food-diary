.PHONY: build install install-local push clean local
BASE_NAME = edo-golang-builder
VERSION ?= $(shell git describe --tags --abbrev=0)
IMG_NAME = $(BASE_NAME):$(VERSION)
REGISTRY_ADDR = harbor.infra.yandex.astral-dev.ru/astral-edo/go/
IMG_NAME_REMOTE = $(REGISTRY_ADDR)$(IMG_NAME)

TARGET = $(HOME)/.local/bin
SCRIPT_PATH = $(TARGET)/$(BASE_NAME)-$(VERSION).sh
RUN_CMD = 'docker run -it --rm \
	-v $$(pwd):/mnt \
	-v $(BASE_NAME)-pkg-cache:/go/pkg/mod/ \
	-v $(BASE_NAME)-build-cache:/go/cache/go-build \
	--add-host git.astralnalog.ru:192.168.1.137'
	 
install:
	$(info Скачиваем образ $(IMG_NAME_REMOTE)) 
	$(info Устанавливаем скрипт запуска $(SCRIPT_PATH))
	docker pull $(IMG_NAME_REMOTE)
	mkdir -p $(TARGET)
	echo $(RUN_CMD) $(IMG_NAME_REMOTE) > $(SCRIPT_PATH)
	chmod +x $(SCRIPT_PATH)

install-local:
	$(info Устанавливаем скрипт запуска $(SCRIPT_PATH))
	mkdir -p $(TARGET)
	echo $(RUN_CMD) $(IMG_NAME) > $(SCRIPT_PATH)
	chmod +x $(SCRIPT_PATH)

build: 
	$(info Начинаем сборку образа.)
	docker build \
	--add-host git.astralnalog.ru:192.168.1.137 \
	-t $(IMG_NAME) -f ./Dockerfile .

push:
	$(info Собираемся выполнить:)
	$(info docker build -t $(IMG_NAME) -f ./Dockerfile .)
	$(info docker tag $(IMG_NAME) $(IMG_NAME_REMOTE))
	$(info docker push $(IMG_NAME_REMOTE))
	$(info Enter чтобы продолжить сборку и ЗАПУШИТЬ ОБРАЗ в registry, Ctrl+C - отменить действие)
	read _
	docker build \
	--add-host git.astralnalog.ru:192.168.1.137 \
	-t $(IMG_NAME) -f ./Dockerfile .
	docker tag $(IMG_NAME) $(IMG_NAME_REMOTE)
	docker push $(IMG_NAME_REMOTE)

clean:
	$(info Удаляем образ и скрипт запуска.)
	rm $(SCRIPT_PATH)
	docker rmi $(IMG_NAME)
	
local: build install-local
