FROM alpine:3.5

COPY bin/swan /swan
WORKDIR /

ENTRYPOINT ["/swan"]
