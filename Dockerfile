# Taken initially from https://www.callicoder.com/docker-golang-image-container-example/
# Because of course free software uses free examples

ARG GOLANG_LATEST="golang@sha256:91b3c5472d9a2ef12f3165aa8979825a5d8b059720b00412f89fc465a04aaa0c"
ARG GOLANG_ALPINE3_13="golang@sha256:4919b2f118f75395f69adcb899c1b796c4337d9649f0ef73aab34eef040149cf"
# Start from the latest golang base image 
# FROM golang@sha256:91b3c5472d9a2ef12f3165aa8979825a5d8b059720b00412f89fc465a04aaa0c as builder # golang:latest
# This next is golang:alpine3.13
FROM ${GOLANG_ALPINE3_13} as builder

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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' -a \
    -o /app/main .
# RUN chown ${USER}:${USER} /app/main

# 2: Time to make it tiny
FROM scratch

# Import from builder for non-root running
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Include the program
COPY --from=builder /app/main /app/main

# Run as not root
USER ${USER}:${USER}

# Command to run the executable
ENTRYPOINT ["/app/main"]