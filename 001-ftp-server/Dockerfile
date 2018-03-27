FROM golang:1.8

ADD ./myftp.go /go/src

WORKDIR /go/src

RUN go build /go/src/myftp.go \
    && rm -rf /go/src/myftp.go 

CMD ["./myftp"]
