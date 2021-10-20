FROM golang:alpine
COPY . ./src

RUN cd ./src && \
    go build main.go


FROM alpine
COPY --from=0 /go/src/main .
CMD ["./main"]
