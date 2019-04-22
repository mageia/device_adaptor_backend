# Build Stage
FROM golang:alpine3.8 AS build-stage
WORKDIR /go/src/device_adaptor/
COPY . .
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add gcc musl-dev
RUN go build -tags=jsoniter -o device_adaptor cmd/main.go

# Final Stage
FROM alpine:3.8
ENV TZ=Asia/Shanghai
WORKDIR /device_adaptor/
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk --no-cache --update add tzdata
COPY --from=build-stage /go/src/device_adaptor/device_adaptor .
COPY opc_alpine /usr/local/bin/opc
CMD ["./device_adaptor"]

EXPOSE 80
VOLUME /device_adaptor/

