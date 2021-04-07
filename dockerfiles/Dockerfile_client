FROM golang:1.16 as ld-client-builder
ARG VERSION
ENV GO111MODULE=on \
        CGO_ENABLED=0 \
        GOOS=linux \
        GOARCH=amd64
WORKDIR /build
COPY ./client .
RUN go mod download
RUN go build -o ld-client -ldflags="-X 'main.Version=$VERSION'" .


FROM alpine
COPY --from=ld-client-builder /build/ld-client /ld-client
ENTRYPOINT ["/ld-client"]
