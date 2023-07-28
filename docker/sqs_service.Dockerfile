FROM golang:alpine

RUN apk --update add ca-certificates git

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main ./cmd/main.go

CMD ["/app/main"]