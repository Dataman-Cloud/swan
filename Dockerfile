FROM alpine:3.5

COPY bin/swan /opt/swan
WORKDIR /

ENTRYPOINT ["/opt/swan"]
