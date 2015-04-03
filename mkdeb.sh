# build for debian based systems
go build -a -o statsd
mkdir -p opt/statsd
mkdir -p etc/statsd
mv statsd opt/statsd/statsd

# build the debian package
fpm -s dir -t deb --name statsd -v $(date +%s).0 etc opt

# clean up
rm -r opt etc
