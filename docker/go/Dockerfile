FROM golang:1.14.7
WORKDIR /go/src/ne-sdk-go-example
COPY ./app .

RUN openssl req -new -newkey rsa:4096 -days 365 -nodes -x509 \
     -subj "/CN=localhost" \
     -keyout srv.key  -out srv.cert

RUN go build

CMD ["./app"]