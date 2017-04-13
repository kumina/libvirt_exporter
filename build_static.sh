#!/bin/sh

docker run -i -v `pwd`:/libvirt_exporter alpine:edge /bin/sh << 'EOF'
set -ex

# Install prerequisites for the build process.
apk update
apk add ca-certificates g++ git go libnl-dev linux-headers make perl pkgconf libtirpc-dev wget
update-ca-certificates

# Install libxml2. Alpine's version does not ship with a static library.
cd /tmp
wget ftp://xmlsoft.org/libxml2/libxml2-2.9.4.tar.gz
tar -xf libxml2-2.9.4.tar.gz
cd libxml2-2.9.4
./configure --disable-shared --enable-static
make -j2
make install

# Install libvirt. Alpine's version does not ship with a static library.
cd /tmp
wget https://libvirt.org/sources/libvirt-3.2.0.tar.xz
tar -xf libvirt-3.2.0.tar.xz
cd libvirt-3.2.0
./configure --disable-shared --enable-static --localstatedir=/var --without-storage-mpath
make -j2
make install
sed -i 's/^Libs:.*/& -lnl -ltirpc -lxml2/' /usr/local/lib/pkgconfig/libvirt.pc

# Build the libvirt_exporter.
cd /libvirt_exporter
export GOPATH=/gopath
go get -d ./...
go build --ldflags '-extldflags "-static"'
strip libvirt_exporter
EOF
