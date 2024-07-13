# CWD := $(strip $(CURDIR))#win
CWD = $(shell pwd)#lin
BUILDER = docker run -it --rm \
	-v $(CWD):/mnt \
	--add-host git.astralnalog.ru:192.168.1.137 \
	harbor.infra.yandex.astral-dev.ru/astral-edo/go/edo-golang-builder:v2.0.6
SERVICE:=food-diary

lint:
	$(BUILDER) golangci-lint run --config=./golangci-lint.yml --timeout=5m --fix

gen-envs:
	$(BUILDER) conf2env -struct Config -file config/config.go -out config/local.env

generate-proto-and-swagger:
	$(BUILDER)  protoc --proto_path proto/$(SERVICE) --proto_path=proto/vendor  \
		--go_out=pkg/$(SERVICE)/  --go_opt=paths=source_relative \
		--plugin=protoc-gen-go=/go/bin/protoc-gen-go \
		--go-grpc_out=pkg/$(SERVICE) --go-grpc_opt=paths=source_relative \
		--plugin=protoc-gen-go-grpc=/go/bin/protoc-gen-go-grpc \
		--grpc-gateway_out=pkg/$(SERVICE) --grpc-gateway_opt=paths=source_relative \
		--plugin=protoc-gen-grpc-gateway=/go/bin/protoc-gen-grpc-gateway \
		--openapiv2_out=allow_merge=true,merge_file_name=api:docs/swagger \
		--plugin=protoc-gen-openapiv2=/go/bin/protoc-gen-openapiv2 \
		proto/$(SERVICE)/package_creator_service.proto
