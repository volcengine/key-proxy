FROM golang:1.18.2-alpine AS dep

WORKDIR /go/src
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main .

FROM alpine:latest
COPY --from=dep /main /main
ENTRYPOINT [ "/main", "-conf-file", "/config.yml"]