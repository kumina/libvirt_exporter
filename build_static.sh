#!/bin/sh

docker run -i -v `pwd`:/libvirt_exporter alpine:3.8 /bin/sh << 'EOF'
set -ex

# Install prerequisites for the build process.
apk update
apk add ca-certificates g++ git go libnl-dev linux-headers make perl pkgconf libtirpc-dev wget libxslt python python-dev
update-ca-certificates

# Install libxml2. Alpine's version does not ship with a static library.
cd /tmp
wget ftp://xmlsoft.org/libxml2/libxml2-2.9.8.tar.gz
tar -xf libxml2-2.9.8.tar.gz
cd libxml2-2.9.8
./configure --disable-shared --enable-static
make -j$(nproc)
make install

# Install libvirt. Alpine's version does not ship with a static library.
cd /tmp
wget https://libvirt.org/sources/libvirt-3.8.0.tar.xz
tar -xf libvirt-3.8.0.tar.xz
cd libvirt-3.8.0
./configure --disable-shared --enable-static --localstatedir=/var --without-storage-mpath
make -j$(nproc)
make install
sed -i 's/^Libs:.*/& -lnl -ltirpc -lxml2/' /usr/local/lib/pkgconfig/libvirt.pc

# Build the libvirt_exporter.
cd /libvirt_exporter
export GOPATH=/gopath
go get -d ./...
go build --ldflags '-extldflags "-static"'
strip libvirt_exporter
EOF
