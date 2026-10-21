# build stage
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.26-alpine AS build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN apk --no-cache add --update ca-certificates && update-ca-certificates

COPY go.mod go.sum /app/

RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
  CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -ldflags="-w -s" -o ./bin/pb-go-api-starter ./cmd/app

# final stage
FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/bin/pb-go-api-starter .

EXPOSE 8000

CMD ["/app/pb-go-api-starter"]
