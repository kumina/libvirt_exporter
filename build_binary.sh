#!/bin/sh

docker run -i -v `pwd`:/gopath/src/github.com/kumina/libvirt_exporter alpine:3.10.5 /bin/sh << 'EOF'
set -ex

# Install prerequisites for the build process.
apk update
apk add ca-certificates gcc g++ git go libnl-dev make perl pkgconf \
libnl3-dev libxml2-dev libxslt-dev libtasn1-dev libvirt-dev
update-ca-certificates

cd /gopath/src/github.com/kumina/libvirt_exporter
go get -d ./...
go build 
strip libvirt_exporter
EOF