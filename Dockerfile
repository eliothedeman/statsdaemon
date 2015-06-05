FROM golang:1.4

ENV DEBIAN_FRONTEND noninteractive

ADD conf.json.example /etc/statsd/conf.json
# build the command
RUN go get -u github.com/eliothedeman/statsdaemon
RUN go build -o /go/bin/statsd github.com/eliothedeman/statsdaemon 

EXPOSE 8125

RUN export PATH=$PATH:/go/bin



ENTRYPOINT /go/bin/statsd -conf=/etc/statsd/conf.json

