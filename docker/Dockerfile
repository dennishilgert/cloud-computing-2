FROM golang:1.21.5-bullseye

WORKDIR /app

# Copy the entire project into the container
COPY . .

# Download dependencies
RUN go mod download

RUN go build -o translator ./cmd

EXPOSE 80

CMD [ "/app/translator" ]