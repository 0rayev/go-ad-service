FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /ad_service ./cmd/app

EXPOSE 8080

CMD [ "/ad_service" ]
