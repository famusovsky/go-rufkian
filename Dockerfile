FROM golang:1.23.1-alpine

WORKDIR /

ARG db_with_ssl=true
ARG port=8080
ARG service=companion

ENV DB_WITH_SSL=$db_with_ssl
ENV PORT=$port

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o out ./cmd/$service

EXPOSE ${port}

CMD [ "sh", "-c", "./out -db_with_ssl=$DB_WITH_SSL -addr=:$PORT" ]
