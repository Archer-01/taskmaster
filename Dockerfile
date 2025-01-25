FROM golang:1.23.4

WORKDIR /app

COPY . .
RUN go build -o /usr/local/bin/taskmasterd cmd/server/main.go
RUN go build -o /usr/local/bin/taskmasterctl cmd/client/main.go

CMD [ "taskmasterd" ]
