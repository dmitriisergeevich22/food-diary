FROM golang:1.22.2-bookworm as prep

# set Go specific envs
ENV GOCACHE="/go/cache/go-build"
ENV GONOPROXY="*.astralnalog.ru"
ENV GONOSUMDB="*.astralnalog.ru" 
ENV GOPRIVATE="*.astralnalog.ru"

# install additional tools
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.18.0 
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3 
RUN go install git.astralnalog.ru/utils/aconf/v3/cmd/conf2env@v3.0.0

RUN curl -sSfL \
  https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
  | sh -s -- -b $(go env GOPATH)/bin v1.56.2 

# install protobuf tools and openapiv2 tools
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.19.1
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.19.1
RUN go install github.com/rakyll/statik@v0.1.7
RUN go install go.uber.org/nilaway/cmd/nilaway@v0.0.0-20240326230412-29def5ee4fba


FROM golang:1.22.2-bookworm

# updating base system to latest
RUN apt-get update 

# installing additions to base system
RUN apt-get install -y --no-install-recommends \
  libxml2-dev \
  libsqlite3-dev \
  protobuf-compiler  \
  libprotobuf-dev \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/* 

# set Go specific envs
ENV GOCACHE="/go/cache/go-build"
ENV GONOPROXY="*.astralnalog.ru"
ENV GONOSUMDB="*.astralnalog.ru" 
ENV GOPRIVATE="*.astralnalog.ru"
ENV GITLAB_AUTH="gitlab+deploy-token-6:CyFgnh8z2eiy2Pcghc4P"

COPY --from=prep /go/bin/* /go/bin/
COPY entrypoint.sh /usr/local/bin/entrypoint.sh

WORKDIR /mnt
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/bin/bash"]
