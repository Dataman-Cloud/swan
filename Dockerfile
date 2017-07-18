FROM daocloud.io/library/golang:1.8.1
WORKDIR /go/src/github.com/Dataman-Cloud/swan
COPY . .
RUN make clean && make

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY bin/swan /opt/swan
WORKDIR /root/
COPY --from=0 /go/src/github.com/Dataman-Cloud/swan/bin/swan .
CMD ["./swan"]
