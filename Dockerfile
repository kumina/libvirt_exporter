# Stage 1: Build libvirt exporter
FROM golang:1.14-alpine3.11 as builder

WORKDIR /usr/src/libvirt_exporter

# Build app
COPY . .
RUN dockerscripts/build_static.sh

# Stage 2: Prepare final image
FROM scratch

# Copy binary from Stage 1
COPY --from=builder libvirt_exporter .

# Entrypoint for starting exporter
ENTRYPOINT [ "./libvirt_exporter" ]
