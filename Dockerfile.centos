FROM centos:7

ENV VERSION 1.8.3
ENV FILE go$VERSION.linux-amd64.tar.gz
ENV URL https://storage.googleapis.com/golang/$FILE
ENV SHA256 1862f4c3d3907e59b04a757cfda0ea7aa9ef39274af99a784f5be843c80c6772

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN set -eux &&\
  yum -y install git make &&\
  yum -y clean all &&\
  curl -OL $URL &&\
	echo "$SHA256  $FILE" | sha256sum -c - &&\
	tar -C /usr/local -xzf $FILE &&\
	rm $FILE &&\
  mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

WORKDIR /go/src/github.com/Dataman-Cloud/swan
COPY . .
RUN make clean && make

FROM centos:7
WORKDIR /root/
COPY --from=0 /go/src/github.com/Dataman-Cloud/swan/bin/swan .
ENTRYPOINT ["./swan"]
