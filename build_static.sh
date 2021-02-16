#!/bin/sh

docker run -i -v `pwd`:/gopath/src/github.com/kumina/libvirt_exporter alpine:3.13.1 /bin/sh << 'EOF'
set -ex

# Install prerequisites for the build process.
apk update
apk add ca-certificates gcc g++ git go libnl-dev linux-headers make perl pkgconf \
libtirpc-dev wget gnutls-dev libxslt-dev python3-dev xen-dev libxml2-utils libpciaccess-dev \
libxml2-dev netcf-dev libnl3-dev libxml2-dev libxslt-dev libtasn1-dev
apk --update add go
update-ca-certificates

# Install libxml2. Alpine's version does not ship with a static library.
#cd /tmp
#wget ftp://xmlsoft.org/libxml2/libxml2-2.9.10.tar.gz
#tar -xf libxml2-2.9.10.tar.gz
#cd libxml2-2.9.10
#./configure --disable-shared --enable-static
#make -j2
#make install

# Install libvirt. Alpine's version does not ship with a static library.
cd /tmp
wget https://libvirt.org/sources/libvirt-5.0.0.tar.xz
tar -xf libvirt-5.0.0.tar.xz
cd libvirt-5.0.0
CFLAGS="-fPIC"
# ./configure --disable-shared --enable-static --localstatedir=/var  \
./configure \
--without-storage-mpath \
--without-storage-scsi \
--without-storage-iscsi \
--without-storage-fs 
make 
make install

sed -i 's/^Libs:.*/& -lnl -ltirpc -lxml2/' /usr/local/lib/pkgconfig/libvirt.pc

# Build the libvirt_exporter.
cd /gopath/src/github.com/kumina/libvirt_exporter
go get -d ./...
go build
strip libvirt_exporter
EOF
