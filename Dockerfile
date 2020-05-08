# Build Stage
FROM golang:alpine AS build-stage
ENV GOPROXY https://goproxy.cn

WORKDIR /go/src/device_adaptor/
ADD go.mod .
ADD go.sum .
RUN go mod download

COPY . .
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add gcc musl-dev
RUN go build -tags=jsoniter -o device_adaptor cmd/main.go

# Final Stage
FROM alpine
WORKDIR /device_adaptor/
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
	&& apk --no-cache --update add tzdata
COPY --from=build-stage /go/src/device_adaptor/device_adaptor .
CMD ["./device_adaptor"]

EXPOSE 80
VOLUME /device_adaptor/

