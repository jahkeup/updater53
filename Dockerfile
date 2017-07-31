FROM golang:1.8 as builder

WORKDIR /go/src/github.com/jahkeup/updater53
COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o updater53


FROM scratch
COPY --from=builder /go/src/github.com/jahkeup/updater53/updater53 /usr/bin/updater53
ENTRYPOINT ["/usr/bin/updater53"]