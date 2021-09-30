FROM golang:1.17-alpine AS build

WORKDIR /hygieia-docker-collector

COPY go.mod go.mod
RUN go mod download

COPY main.go main.go
COPY internal internal
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

# tiny run container
FROM alpine:latest
COPY --from=build /hygieia-docker-collector/app /hygieia-docker-collector/app

CMD /hygieia-docker-collector/app
