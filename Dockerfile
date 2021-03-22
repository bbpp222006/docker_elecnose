FROM golang:alpine
COPY main.go .
ENV GOPROXY https://goproxy.io


RUN apk update && apk add --no-cache git
RUN go get -d -v ./... 
RUN go build main.go

FROM alpine
COPY --from=0 /go/main .
CMD ["./main"]