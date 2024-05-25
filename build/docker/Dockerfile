# use specific versions of images
FROM openapitools/openapi-generator-cli:v7.0.1 as api

WORKDIR /local

COPY api api
COPY scripts scripts
COPY .openapi-generator-ignore .openapi-generator-ignore

RUN docker-entrypoint.sh generate -c /local/scripts/generator-cfg.yaml

# not used normally but redefine entrypoint for the case of checking this stage results
ENTRYPOINT ["bash"]

############################################

############################################
FROM golang:latest AS build

WORKDIR /app

# download dependencies - low frequency of changes
COPY go.mod go.sum ./
RUN go mod download

# copy sources - higher frequency of changes
COPY internal/ internal/
COPY cmd/ cmd/
COPY --from=api /local/ ./

# ensure tests are passing
RUN go test ./...

# create executable - sprava-krvi-webapi-srv
# we want to use scratch image so setting
# the build options in the way that will link all dependencies statically
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
      -ldflags="-w -s" \
      -installsuffix 'static' \
      -o ./sprava-krvi-webapi-srv ./cmd/sprava-krvi-api-service

############################################

FROM scratch

# see https://github.com/opencontainers/image-spec/blob/main/annotations.md for details
LABEL org.opencontainers.image.authors="Samuel Schmidt, Marek Sykora"
LABEL org.opencontainers.image.title="Blood management WebAPI Service"
LABEL org.opencontainers.image.description="WebAPI for managing blood donors and donated blood units"

# list all variables and their default values for clarity
ENV API_ENVIRONMENT=production
ENV API_PORT=8080
ENV API_MONGODB_HOST=mongo
ENV API_MONGODB_PORT=27017
ENV API_MONGODB_DATABASE=ss-sprava-krvi
# ENV API_MONGODB_COLLECTION=<collection>
ENV API_MONGODB_USERNAME=root
ENV API_MONGODB_PASSWORD=neUhaDnes
ENV API_MONGODB_TIMEOUT_SECONDS=5

COPY --from=build /app/sprava-krvi-webapi-srv ./

# Actual port may be changed during runtime
# Default using for the simple case scenario
EXPOSE 8080
ENTRYPOINT ["./sprava-krvi-webapi-srv"]
