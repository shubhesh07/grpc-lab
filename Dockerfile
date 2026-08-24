# grpcurl is bundled, so the image is the only thing to install.
FROM golang:1.25-alpine AS build
RUN go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags '-s -w' -o /grpc-lab .

FROM alpine:3.20
COPY --from=build /go/bin/grpcurl /grpc-lab /usr/local/bin/
WORKDIR /data
EXPOSE 8090
# Bind to all interfaces inside the container; docker -p keeps it local.
ENTRYPOINT ["grpc-lab", "-bind", "0.0.0.0"]
