FROM microservice

RUN apt-get update && apt-get install -y --no-install-recommends \
		g++ \
		gcc \
		git \
		libc6-dev \
		make \
		rpm \
		ruby-dev \
	&& rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.6.2
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 e40c36ae71756198478624ed1bb4ce17597b3c19d243f3f0899bb5740d56212a

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
	&& echo "$GOLANG_DOWNLOAD_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz

ENV GOPATH /go
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin"
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN go get github.com/tools/godep

# install package builder
RUN gem install fpm

ADD ./ /opt/armada-stats
#add to go workspace
RUN mkdir -p "$GOPATH/src/github.com/krise3k"
RUN ln -s /opt/armada-stats "$GOPATH/src/github.com/krise3k/"


#install go dependencies
WORKDIR "$GOPATH/src/github.com/krise3k/armada-stats"

RUN cd "$GOPATH/src/github.com/krise3k/armada-stats" && godep restore && go build .

ADD ./supervisor/*.conf /etc/supervisor/conf.d/

VOLUME ["/var/run/docker.sock"]

EXPOSE 80