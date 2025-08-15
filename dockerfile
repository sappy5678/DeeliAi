FROM golang:1.25.0
RUN apt update && apt install jq -y
WORKDIR /app

COPY go.* /app/
RUN go mod download

COPY . . 

RUN make build

CMD ["make", "run"]