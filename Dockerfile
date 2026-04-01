FROM golang:alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o network_exporter .

FROM alpine:latest
RUN apk --no-cache add ca-certificates libcap iproute2 && mkdir -p /app/cfg
WORKDIR /app
COPY --from=builder /build/network_exporter network_exporter
RUN setcap 'cap_net_raw,cap_net_admin+eip' /app/network_exporter
CMD ["/app/network_exporter"]
EXPOSE 9427
