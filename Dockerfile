# Build Stage
FROM golang:alpine3.8 AS build-stage
WORKDIR /go/src/deviceAdaptor/
COPY . .
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add gcc musl-dev
RUN go build -tags=jsoniter -o server cmd/main.go

# Final Stage
FROM alpine:3.8
ENV TZ=Asia/Shanghai
WORKDIR /deviceAdaptor/
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk --no-cache --update add tzdata
COPY --from=build-stage /go/src/deviceAdaptor/server .
CMD ["./server"]
EXPOSE 80

