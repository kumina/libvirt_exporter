# Stage 1: Build libvirt exporter
FROM golang:alpine

# Install dependencies
RUN apk add --update git gcc g++ make libc-dev portablexdr-dev linux-headers libnl-dev perl libtirpc-dev pkgconfig wget
RUN wget ftp://xmlsoft.org/libxml2/libxml2-2.9.4.tar.gz -P /tmp && \
    tar -xf /tmp/libxml2-2.9.4.tar.gz -C /tmp
WORKDIR /tmp/libxml2-2.9.4
RUN ./configure --disable-shared --enable-static && \
    make -j2 && \
    make install
RUN wget https://libvirt.org/sources/libvirt-3.2.0.tar.xz -P /tmp && \
    tar -xf /tmp/libvirt-3.2.0.tar.xz -C /tmp
WORKDIR /tmp/libvirt-3.2.0
RUN ./configure --disable-shared --enable-static --localstatedir=/var --without-storage-mpath && \
    make -j2 && \
    make install && \
    sed -i 's/^Libs:.*/& -lnl -ltirpc -lxml2/' /usr/local/lib/pkgconfig/libvirt.pc

# Prepare working directory
ENV LIBVIRT_EXPORTER_PATH=/go/src/github.com/kumina/libvirt_exporter
RUN mkdir -p $LIBVIRT_EXPORTER_PATH
WORKDIR $LIBVIRT_EXPORTER_PATH
COPY . .

# Build and strip exporter
RUN go get -d ./... && \
    go build --ldflags '-extldflags "-static"' && \
    strip libvirt_exporter

# Stage 2: Prepare final image
FROM scratch

# Copy binary from Stage 1
COPY --from=0 /go/src/github.com/kumina/libvirt_exporter/libvirt_exporter .

# Entrypoint for starting exporter
ENTRYPOINT [ "./libvirt_exporter" ]
