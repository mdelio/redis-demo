# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/mdelio/redis-demo

# Build the redis_demo command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/go-redis/redis
RUN go install github.com/mdelio/redis-demo

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/redis-demo  -listen_addr :80

# Document that the service listens on port 80.
EXPOSE 80