BUILD=go build -a
CWD=$(shell pwd)
IMAGE_NAME=statsd-go

make:
	go generate ./...
	$(BUILD) -o bin/statsd github.com/eliothedeman/statsdaemon
	
install:
	go get -u github.com/eliothedeman/statsdaemon
	go install -a github.com/eliothedeman/statsdaemon

build:
	docker build -t $(IMAGE_NAME) .

start: 
	cwd=$(pwd)
	docker run -v $(CWD)/conf.json:/etc/statsd/conf.json -p 8125:8125/udp --name $(IMAGE_NAME) -d $(IMAGE_NAME)

kill:
	docker kill $(IMAGE_NAME)

clean: kill
	docker rm $(IMAGE_NAME)

deb:
	# build for debian based systems
	go build -a -o statsd
	mkdir -p opt/statsd
	mkdir -p etc/statsd
	mv statsd opt/statsd/statsd

	# build the debian package
	fpm -s dir -t deb --name statsd -v $(date +%s).0 etc opt

	# clean up
	rm -r opt etc


test:
	go test -p=1 ./...
