FROM golang:1.12

# TODO: Release tracking?
# ARG VERSION="dirty"
# -ldflags "-X github.com/tokend/stellar-withdraw-svc/config.Release=${VERSION}" 

WORKDIR /go/src/github.com/tokend/stellar-withdraw-svc
COPY . .
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build -o /usr/local/bin/stellar-withdraw-svc github.com/tokend/stellar-withdraw-svc

###

FROM alpine:3.9

COPY --from=0 /usr/local/bin/stellar-withdraw-svc /usr/local/bin/stellar-withdraw-svc
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["stellar-withdraw-svc", "run", "withdraw"]

