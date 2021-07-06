# Taken initially from https://www.callicoder.com/docker-golang-image-container-example/
# Because of course free software uses free examples

# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="James Fisher <jamesfisher314@outlook.com>"

# Do things so the container is more secure
RUN apt-get update && apt-get install git ca-certificates tzdata && update-ca-certificates && apt-get clean
ENV USER=appuser
ENV UID=31415
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/dev/null" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
# COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN  go get -d -v

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Run as not root
USER ${USER}

# Command to run the executable
CMD ["./main"]