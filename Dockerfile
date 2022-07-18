FROM golang:1.18 AS builder

ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/zostay/ext-authz-keep-out
COPY . .

RUN go install ./
RUN ls /go/bin

CMD [ "/go/bin/ext-authz-keep-out" ]

FROM scratch AS app

COPY --from=builder /go/bin/ext-authz-keep-out /ext-authz-keep-out

ENTRYPOINT [ "/ext-authz-keep-out" ]
