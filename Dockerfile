FROM golang:1.8 as builder

WORKDIR /go/src/github.com/jahkeup/updater53
COPY . .

ARG GOOS=linux
ARG GOARCH=amd64
ARG GOARM=

RUN go get -d -v ./...
RUN CGO_ENABLED=0 go build -v -a -installsuffix cgo -o updater53


FROM scratch
COPY --from=builder /go/src/github.com/jahkeup/updater53/updater53 /usr/bin/updater53
ENTRYPOINT ["/usr/bin/updater53"]