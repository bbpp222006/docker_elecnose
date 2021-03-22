FROM golang:alpine
COPY main.go .
ENV GOPROXY https://goproxy.io

RUN apk add --no-cache ca-certificates git && \
    go get -d -v ./... && \
    go build main.go

FROM alpine
COPY --from=0 /go/main .
CMD ["./main"]