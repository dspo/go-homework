FROM docker.cnb.cool/dspo-group/go-example2/ubuntu-go:latest AS builder

WORKDIR /go/src/app

COPY . .

ENV GOPROXY=https://goproxy.cn,direct

RUN go mod tidy
RUN go build -o app cmd/app/main.go

FROM docker.cnb.cool/dspo-group/go-example2/ubuntu:latest
COPY --from=builder /go/src/app/app /app

EXPOSE 8080

ENTRYPOINT ["/app"]
