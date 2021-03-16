FROM golang:1.16 as builder
ENV GO111MODULE=on \
        CGO_ENABLED=0 \
        GOOS=linux \
        GOARCH=amd64
WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o ld .


FROM scratch
COPY --from=builder /build/ld /
COPY --from=builder /proto/ld.proto /ld.proto
ENV PORT 5326
EXPOSE 5326
ENTRYPOINT ["/ld"]