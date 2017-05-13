# Armada-stats
It is a running daemon that collects, aggregates, processes, and exports to [influxdb](https://influxdata.com/) information about running [armada](http://armada.sh) containers.

### Collected data
#### service
	ID           ID
	service      name
	addres       listen address
	status       container status [0 - passing, 1 - warning, 2 - critical]
	status_name  name of service status it's mapped to status
	tags         container tags
	uptime       uptime in seconds
	host

    cpu_percentage      average CPU usage ie. if host has 16 cores, max CPU usage will be 1600
    cpu_core_percentage average CPU per core usage, in same case as above, max CPU per core will be 100
    memory              memory usage, without cache, in bytes
    memory_limit        momory limit in bytes
    memory_percentage   percent memory usage
    swap                swap usage in bytes
    network_rx          total number of network Rx in bytes
    network_tx          total number of network Tx in bytes
    block_read          total number of BlockRead in bytes
    block_write         total number of BlockWrite in bytes

#### ship
- total number of services grouped by `status_name`
    
### How to start developing
- Start vagrant `vagrant up`
- Log into it `vagrant ssh`
- Run influx `armada run influxdb -r armada-stats-influxdb -v '<local dir>:/var/influxdb'`
- Create a *custom.yml* in the conf directory to override default configuration options. Especially *armada_host*
- Build container `cd /projects/armada-stats && armada build`
- Run container `armada run --env dev -v /var/run/docker.sock:/var/run/docker.sock`
- Log into `armada ssh`
- Build armada-stats `go build .`
- Restart armada-stats `supervisorctl restart armada-stats`

To see changes after developing run `go build . && supervisorctl restart armada-stats`

### How to build package
`./packaging/build_package.sh`

