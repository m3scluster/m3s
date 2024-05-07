FROM golang:alpine AS builder

WORKDIR /build

COPY . /build/

RUN apk add git && \
    go get -d

ARG TAG
ARG BUILDDATE
ARG VERSION_URL
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=$BUILDDATE -X main.GitVersion=$TAG -X main.VersionURL=$VERSION_URL -extldflags \"-static\"" -o main .


FROM alpine:3.19
LABEL maintainer="Andreas Peters <support@aventer.biz>"
LABEL org.opencontainers.image.title="mesos-m3s" 
LABEL org.opencontainers.image.description="ClusterD/Apache Mesos framework to run Kubernetes"
LABEL org.opencontainers.image.vendor="AVENTER UG (haftungsbeschränkt)"
LABEL org.opencontainers.image.source="https://github.com/AVENTER-UG/"

ENV DOCKER_RUNNING=true

RUN apk add --no-cache ca-certificates
RUN adduser -S -D -H -h /app appuser
USER appuser

COPY --from=builder /build/main /app/

EXPOSE 10000

WORKDIR "/app"

CMD ["./main"]
