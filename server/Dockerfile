FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o kvserver

FROM alpine
COPY --from=builder /app/kvserver /kvserver
EXPOSE 8080
CMD ["/kvserver"]

