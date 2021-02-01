FROM alpine:3.5
LABEL description="A simple http sink"

# Install some basic helper tools for debugging
RUN apk add --no-cache bash tzdata ca-certificates

WORKDIR /app
COPY build/mockpost /app/
COPY configs/ /app/configs
COPY build/key.pem /app/
COPY build/cert.pem /app/

# will require a --config flag at runtime, and likely GOMAXPROCS
ENTRYPOINT ["/app/mockpost"]

EXPOSE 443
