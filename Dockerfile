FROM golang:1.8 as builder

WORKDIR /go/src/github.com/jahkeup/updater53
COPY . .

ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

RUN go get -d -v ./...
RUN CGO_ENABLED=0 go build -v -a -installsuffix cgo -o updater53

###############################################################################

# Taken from https://github.com/aws/amazon-ecs-agent/blob/ecda8a686200643081fe7de498b61b1c023b2c25/misc/certs/Dockerfile
FROM debian:latest as certs

RUN apt-get update &&  \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# If anyone has a better idea for how to trim undesired certs or a better ca list to use, I'm all ears
RUN cp /etc/ca-certificates.conf /tmp/caconf && cat /tmp/caconf | \
  grep -v "mozilla/CNNIC_ROOT\.crt" > /etc/ca-certificates.conf && \
update-ca-certificates --fresh

###############################################################################

FROM scratch
COPY --from=builder /go/src/github.com/jahkeup/updater53/updater53 /usr/bin/updater53
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/usr/bin/updater53"]