FROM golang:latest as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o consul-bigip-tg .

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.1
COPY --from=builder /build/consul-bigip-tg /app/
RUN echo $'[consul]\n[bigip]\n[gateway]' >> /app/config.toml
WORKDIR /app
USER 1001

CMD ./consul-bigip-tg
