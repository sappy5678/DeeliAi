FROM golang:1.25.0
RUN apt update && apt install jq -y
WORKDIR /app

COPY go.* /app/
RUN go mod tidy && go mod vendor

COPY . . 

RUN make build

CMD ["make", "run"]