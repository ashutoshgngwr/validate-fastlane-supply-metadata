FROM golang:1.14-alpine as builder
RUN apk add --no-cache -q binutils
WORKDIR /app
ADD ./ /app
RUN go build -ldflags '-extldflags "-static"' -a -o /entrypoint . && \
    strip /entrypoint

FROM scratch
COPY --from="builder" /entrypoint /entrypoint
ENTRYPOINT ["/entrypoint"]
