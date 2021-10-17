FROM golang:alpine as builder

WORKDIR /build

COPY . /build/

RUN apk add git && \
    go get -d

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.MinVersion=`date -u +%Y%m%d%.H%M%S` -extldflags \"-static\"" -o main .


FROM alpine
LABEL maintainer="Andreas Peters <support@aventer.biz>"

RUN apk add --no-cache ca-certificates
RUN adduser -S -D -H -h /app appuser

USER appuser

COPY --from=builder /build/main /app/

EXPOSE 10000

WORKDIR "/app"

CMD ["./main"]
