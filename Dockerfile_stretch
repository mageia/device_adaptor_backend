# Final Stage
FROM golang:1.12-stretch
ENV TZ=Asia/Shanghai
WORKDIR /go/device_adaptor/
COPY . .
RUN go build -tags=jsoniter -o device_adaptor cmd/main.go
COPY opc /usr/local/bin
CMD ["./cmd/device_adaptor"]

EXPOSE 80
VOLUME /go/device_adaptor/

