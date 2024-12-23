FROM golang:1.23.4

WORKDIR /app

COPY . .
RUN go build -o /usr/local/bin/taskmasterd cmd/server/main.go

CMD [ "taskmasterd" ]
