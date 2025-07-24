# ---- Stage 1: Build Stockfish for x86_64 ----
FROM ubuntu:22.04 AS stockfish-build

RUN apt-get update && apt-get install -y \
    build-essential \
    git \
    curl

WORKDIR /build
RUN git clone https://github.com/official-stockfish/Stockfish.git
WORKDIR /build/Stockfish

# Checkout Stockfish 17.1 (latest stable release)
RUN git checkout sf_17

WORKDIR /build/Stockfish/src

# Build Stockfish (let it auto-detect the best architecture)
RUN make build

# ---- Stage 2: Build Go server ----
FROM golang:1.21 AS go-build

WORKDIR /app
COPY . .
RUN go build -o chess-stockfish ./cmd

# ---- Stage 3: Final image - use Ubuntu 22.04 to match GLIBC versions ----
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy Go server
COPY --from=go-build /app/chess-stockfish /app/chess-stockfish

# Copy Stockfish binary
COPY --from=stockfish-build /build/Stockfish/src/stockfish /usr/local/bin/stockfish

# Copy web assets
COPY web /app/web

# Expose the port your Go server listens on
EXPOSE 8080

# Entrypoint: run your Go server (which will launch Stockfish as a subprocess)
ENTRYPOINT ["/app/chess-stockfish"] 