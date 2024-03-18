FROM golang:1.21.5-bullseye

WORKDIR /app

# The contents of the /app directory on the host are copied over to the /app directory in the container.
COPY /app .

# Download go packages.
RUN go mod download

RUN go build -o translator ./cmd/main.go

EXPOSE 80

CMD [ "/app/translator" ]