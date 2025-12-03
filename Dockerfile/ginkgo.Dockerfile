FROM docker.cnb.cool/dspo-group/go-example2/ubuntu-go:latest AS builder

WORKDIR /go/src/app

COPY . .

RUN mkdir "bin"
ENV GOPROXY=https://goproxy.cn,direct
ENV GOBIN=/go/src/app/bin

# install kubectl
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*
RUN set -e; \
    arch="$(dpkg --print-architecture)"; arch="${arch##*-}"; \
    echo "arch: $arch"; \
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/$arch/kubectl"; \
    mv kubectl bin/kubectl; \
    chmod +x bin/kubectl;

# install ginkgo
RUN go install -v github.com/onsi/ginkgo/v2/ginkgo@latest
RUN chmod +x /go/src/app/bin/ginkgo

WORKDIR /go/src/app/test/conformance
RUN /go/src/app/bin/ginkgo build -r -o conformance.test

WORKDIR /go/src/app/test/e2e
RUN /go/src/app/bin/ginkgo build -r -o e2e.test

FROM docker.cnb.cool/dspo-group/go-example2/ubuntu:latest

WORKDIR /usr/local/huayi

COPY --from=builder /go/src/app/bin/kubectl /usr/local/bin/kubectl

# copy conformance.test here
COPY --from=builder /go/src/app/test/conformance/conformance.test ./

# copy e2e.test here
COPY --from=builder /go/src/app/test/e2e/e2e.test ./

ENV PATH=$PATH:/usr/local/bin

# the defualt ENTRYPOINT is to run e2e.test,
# you can override it at the runtime to run conformance.test
ENTRYPOINT ["/usr/local/huayi/e2e.test"]
CMD ["--ginkgo.trace", "-test.v", "--ginkgo.v", "-test.failfast", "--ginkgo.fail-fast"]
