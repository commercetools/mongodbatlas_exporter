FROM quay.io/prometheus/golang-builder:1.16-base as builder

WORKDIR /go/src/github.com/commercetools/mongodbatlas_exporter
ADD .   /go/src/github.com/commercetools/mongodbatlas_exporter

RUN make build

FROM quay.io/prometheus/busybox:latest

COPY --from=builder /go/src/github.com/commercetools/mongodbatlas_exporter/mongodbatlas_exporter  /bin/mongodbatlas_exporter

EXPOSE 9905
ENTRYPOINT [ "/bin/mongodbatlas_exporter" ]
