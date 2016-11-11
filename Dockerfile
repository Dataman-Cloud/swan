FROM golang:1.6.3-alpine

COPY . /go/src/github.com/Dataman-Cloud/swan
WORKDIR /go/src/github.com/Dataman-Cloud/swan
RUN go build -v  && install -v swan /usr/local/bin
ENTRYPOINT ["/usr/local/bin/swan"]

