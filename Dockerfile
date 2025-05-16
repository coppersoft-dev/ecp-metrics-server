ARG GO_VERSION=1.24.3-alpine@sha256:ef18ee7117463ac1055f5a370ed18b8750f01589f13ea0b48642f5792b234044
ARG XX_VERSION=1.6.1@sha256:923441d7c25f1e2eb5789f82d987693c47b8ed987c4ab3b075d6ed2b5d6779a3

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} as builder

# Copy the build utilities.
COPY --from=xx / /

WORKDIR /workspace

# copy modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache modules
RUN go mod download

# copy source code
COPY main.go main.go
COPY cd cd

# build
ARG TARGETPLATFORM

ENV CGO_ENABLED=0
RUN xx-go build -trimpath -a -o ecp-metrics-server main.go

FROM gcr.io/distroless/static-debian12:nonroot@sha256:c0f429e16b13e583da7e5a6ec20dd656d325d88e6819cafe0adb0828976529dc

COPY --from=builder /workspace/ecp-metrics-server /usr/local/bin/

ENTRYPOINT [ "ecp-metrics-server" ]
