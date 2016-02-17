FROM microservice_go

RUN apt-get update && apt-get install -y --no-install-recommends \
		ruby-dev \
	&& rm -rf /var/lib/apt/lists/*
	
# install package builder
RUN gem install fpm

ADD ./ /opt/armada-stats
#add to go workspace
RUN mkdir -p "$GOPATH/src/github.com/krise3k"
RUN ln -s /opt/armada-stats "$GOPATH/src/github.com/krise3k/"


#install go dependencies
WORKDIR "$GOPATH/src/github.com/krise3k/armada-stats"

RUN godep restore
RUN go build .

ADD ./supervisor/*.conf /etc/supervisor/conf.d/

VOLUME ["/var/run/docker.sock"]

EXPOSE 80