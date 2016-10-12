FROM microservice
ENV LANG='en_US.UTF-8' LANGUAGE='en_US:en' LC_ALL='en_US.UTF-8'
RUN apt-get update && apt-get install -y --no-install-recommends \
		g++ \
		gcc \
		git \
		libc6-dev \
		make \
		rpm \
		ruby-dev \
	&& rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.7.1
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA256 43ad621c9b014cde8db17393dc108378d37bc853aa351a6c74bf6432c1bbd182

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