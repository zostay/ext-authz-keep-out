FROM golang:1.18 AS builder

WORKDIR /go/src/github.com/zostay/ext-authz-keep-out
COPY . .

RUN go install ./

FROM busybox AS app

COPY --from=builder /go/bin/ext-authz-keep-out /usr/local/bin/ext-authz-keep-out

ENTRYPOINT [ "/usr/local/bin/ext-authz-keep-out" ]
