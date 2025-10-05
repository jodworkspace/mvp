FROM golang:1.25 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./mvp

FROM alpine

WORKDIR /server
COPY --from=builder /build/mvp ./mvp

EXPOSE 9731

CMD ["/server/mvp"]

# docker build -t mvp:0.0.1
# docker run -p 9731:9731 mvp:0.0.1