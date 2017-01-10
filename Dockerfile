FROM golang:1.6.3-alpine

COPY . /go/src/github.com/Dataman-Cloud/swan
WORKDIR /go/src/github.com/Dataman-Cloud/swan
RUN go build -ldflags "-X github.com/Dataman-Cloud/swan/srv/version.BuildTime=`date -u +%Y-%m-%d:%H-%M-%S` -X github.com/Dataman-Cloud/swan/src/version.Version=0.011"  -v -o bin/swan main.go node.go

EXPOSE 9999 2111 53 80

ENTRYPOINT ["/go/src/github.com/Dataman-Cloud/swan/bin/swan"]

