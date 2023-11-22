# compile fx-core
FROM golang:1.21.4-alpine3.18 as builder

RUN apk add --no-cache git build-base linux-headers

WORKDIR /app

COPY . .

RUN make build

# build fx-core
FROM alpine:3.18

WORKDIR root

COPY --from=builder /app/build/bin/fxcored /usr/bin/fxcored

EXPOSE 26656/tcp 26657/tcp 26660/tcp 9090/tcp 1317/tcp 8545/tcp 8546/tcp

VOLUME ["/root"]

ENTRYPOINT ["fxcored"]
