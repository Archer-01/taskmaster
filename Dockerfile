FROM golang:1.23.4

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN apt update && apt install -y iputils-ping

COPY . .
RUN go build -o /usr/local/bin/taskmasterd cmd/server/main.go
RUN go build -o /usr/local/bin/taskmasterctl cmd/client/main.go

CMD [ "taskmasterd" ]
# CMD ["tail", "-f"]
