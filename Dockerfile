FROM docker.medzdrav.ru/golang:1.0

WORKDIR /go

COPY src /go/src
COPY db /go/db

#RUN goose up

WORKDIR /go/src/chats

RUN dep ensure

WORKDIR /go/src

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /chats chats/main

FROM alpine:3.7

RUN apk --no-cache add ca-certificates \
    && apk add --no-cache tzdata

WORKDIR /
COPY --from=0 /chats .
CMD ["./chats"]

EXPOSE 8000