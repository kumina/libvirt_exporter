#!/bin/sh
libxml2_ver=2.9.8
libvirt_ver=3.8.0
apk add --update git gcc g++ make libc-dev portablexdr-dev linux-headers libnl-dev perl libtirpc-dev pkgconfig wget python2 python2-dev libxslt
wget ftp://xmlsoft.org/libxml2/libxml2-${libxml2_ver}.tar.gz -P /tmp && \
    tar -xf /tmp/libxml2-${libxml2_ver}.tar.gz -C /tmp
cd /tmp/libxml2-${libxml2_ver}
./configure --disable-shared --enable-static && \
    make -j$(nproc) && \
    make install
wget https://libvirt.org/sources/libvirt-${libvirt_ver}.tar.xz -P /tmp && \
    tar -xf /tmp/libvirt-${libvirt_ver}.tar.xz -C /tmp
mkdir /tmp/libvirt-${libvirt_ver}/build
cd /tmp/libvirt-${libvirt_ver}/build
../configure --disable-shared --enable-static --localstatedir=/var --without-storage-mpath && \
    make -j$(nproc) && \
    make install && \
    sed -i 's/^Libs:.*/& -lnl -ltirpc -lxml2/' /usr/local/lib/pkgconfig/libvirt.pc
cd /usr/src/libvirt_exporter
echo build go binary
go build --ldflags '-extldflags "-static"' -o libvirt_exporter
echo strip go binary
strip libvirt_exporter
