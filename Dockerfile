FROM golang:alpine AS builder

COPY main.go .


RUN apk update && apk add --no-cache git
RUN go get github.com/stianeikeland/go-rpio/v4
RUN go get github.com/imroc/req
RUN go build main.go

FROM alpine
COPY --from=0 /go/main .
CMD ["./main"]