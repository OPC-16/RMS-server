# Use the official Golang image as the base image
FROM golang:1.22

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownload them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# copy the source code
COPY . .

# build the application
RUN go build -v -o rms-server .

# command to run the application
CMD ["./rms-server"]
