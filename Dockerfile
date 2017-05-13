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

ENV GOLANG_VERSION 1.8.1
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 a579ab19d5237e263254f1eac5352efcf1d70b9dacadb6d6bb12b0911ede8994

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

#add to go workspace
RUN mkdir -p "$GOPATH/src/github.com/krise3k"
ADD ./ "$GOPATH/src/github.com/krise3k/armada-stats

#install go dependencies
WORKDIR "$GOPATH/src/github.com/krise3k/armada-stats"

RUN go build .

ADD ./supervisor/*.conf /etc/supervisor/conf.d/

VOLUME ["/var/run/docker.sock"]

EXPOSE 80