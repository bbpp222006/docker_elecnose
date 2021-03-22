FROM golang:alpine AS builder

COPY main.go .
ENV GOPROXY https://goproxy.io


RUN apk update && apk add --no-cache git
RUN go mod download
RUN go build main.go

FROM alpine
COPY --from=0 /go/main .
CMD ["./main"]