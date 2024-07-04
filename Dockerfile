# syntax=docker/dockerfile:1

FROM golang:1.21.1-alpine as builder


WORKDIR /usr/src/app

COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY . .

# Run unit tests
RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/app ./cmd/


FROM scratch
COPY --from=builder /usr/local/bin/app /bin/app

EXPOSE 8080

CMD ["/bin/app"]