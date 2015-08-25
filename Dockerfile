FROM golang:1.5

RUN mkdir -p /go/src
WORKDIR /go/src
COPY . /go/src

RUN GOPATH=/go go build -o httptester ./

VOLUME [/conf]

ENTRYPOINT ["./httptester"]
CMD ["/conf/testsuite.json"]
