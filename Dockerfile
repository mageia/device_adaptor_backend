# Build Stage
FROM golang:alpine3.8 AS build-stage
WORKDIR /go/src/device_adaptor
COPY . .
RUN go build -o server

# Final Stage
FROM alpine:3.8
ENV TZ=Asia/Shanghai

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk --no-cache --update add tzdata
COPY --from=build-stage /go/src/device_adaptor/server . 
CMD ["./server"]

