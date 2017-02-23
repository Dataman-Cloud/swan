FROM alpine:3.5

COPY bin/swan /swan
WORKDIR /

EXPOSE 9999 2111 53 80

ENTRYPOINT ["/swan"]
