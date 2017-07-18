FROM golang:1.8.1-alpine
RUN apk --no-cache add make git
WORKDIR /go/src/github.com/Dataman-Cloud/swan
COPY . .
RUN make clean && make

FROM alpine:3.5
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/Dataman-Cloud/swan/bin/swan .
ENTRYPOINT ["./swan"]
