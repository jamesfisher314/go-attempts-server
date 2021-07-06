# Taken initially from https://www.callicoder.com/docker-golang-image-container-example/
# Because of course free software uses free examples

# Start from the latest golang base image 
# FROM golang@sha256:91b3c5472d9a2ef12f3165aa8979825a5d8b059720b00412f89fc465a04aaa0c as builder # golang:latest
# This next is golang:alpine3.13
FROM golang@sha256:4919b2f118f75395f69adcb899c1b796c4337d9649f0ef73aab34eef040149cf as builder

# Add Maintainer Info
LABEL maintainer="James Fisher <jamesfisher314@outlook.com>"

# Do things so the container is more secure
# RUN apt-get update && apt-get install git ca-certificates tzdata && update-ca-certificates && apt-get clean # golang:latest
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates
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

# Copy the source from the current directory to the Working Directory inside the container, including go mod and sum
COPY . .

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go get -d -v

# Build the Go app
RUN go build -o /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Run as not root
USER ${USER}

# Command to run the executable
CMD ["/app/main"]