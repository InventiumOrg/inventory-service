FROM docker.io/golang:1.24

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./inventory-service .

CMD ["/inventory-service"]