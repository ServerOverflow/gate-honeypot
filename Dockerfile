FROM --platform=$BUILDPLATFORM golang:1.24.3 AS build

WORKDIR /workspace
COPY go.mod go.sum ./

RUN go mod download

COPY plugins ./plugins
COPY util ./util
COPY gate.go ./

ARG TARGETOS TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w" -a -o gate gate.go

FROM --platform=$BUILDPLATFORM debian:bullseye-slim AS app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build /workspace/gate /
CMD ["/gate"]
