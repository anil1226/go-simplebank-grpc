FROM golang:1.22.1 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

FROM alpine:3.19.1
WORKDIR /app
COPY --from=build /app/main .
COPY app.env .

EXPOSE 8080
CMD [ "/app/main" ]