FROM golang:1.23.3-bookworm as build

WORKDIR $GOPATH/app
COPY go.mod go.sum .
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /glpatEye .

# FROM gcr.io/distroless/static-debian12 as final
FROM alpine:3.21.0 as final
RUN addgroup -g 3456 nonroot && adduser -D -G nonroot -u 3456 nonroot
WORKDIR /usr/local/bin/
COPY --from=build --chown=nonroot:nonroot /glpatEye ./glpatEye
USER nonroot:nonroot

CMD ["./glpatEye"]