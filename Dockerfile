# Stage 1: Build the Go application
FROM golang:1.23.1-bookworm AS builder

# Set the working directory inside the container
WORKDIR /app

# Install RocksDB dependencies (Debian-based)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgflags-dev \
    libsnappy-dev \
    zlib1g-dev \
    libbz2-dev \
    liblz4-dev \
    libzstd-dev \
    build-essential \
    wget \
    cmake \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Set the RocksDB version you want to install - 9.6.1 is not yet part of the current Debian distribution
# We need 9.6.1 to be able to use the latest Go bindings
ENV ROCKSDB_VERSION=9.6.1

# Download and compile RocksDB
RUN wget https://github.com/facebook/rocksdb/archive/refs/tags/v$ROCKSDB_VERSION.tar.gz \
    && tar -xzf v$ROCKSDB_VERSION.tar.gz \
    && cd rocksdb-$ROCKSDB_VERSION \
    && make -j$(nproc) shared_lib \
    && make install-shared

# Remove unnecessary files to reduce image size
RUN rm -rf /rocksdb-$ROCKSDB_VERSION v$ROCKSDB_VERSION.tar.gz

# Verify that RocksDB was installed correctly
RUN ldconfig && ldconfig -p | grep rocksdb

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the Go code
COPY . .

# Build the Go application
RUN CGO_CFLAGS="-I/usr/local/include/rocksdb" \
    CGO_LDFLAGS="-L/usr/local/lib -lrocksdb -lstdc++ -lm -lz -lsnappy -llz4 -lzstd" \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o toyapp .

# Stage 2: Create the runnable image
FROM debian:bookworm-slim

# Set the working directory
WORKDIR /app

# Install RocksDB dependencies (Debian-based)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgflags2.2 \
    libsnappy1v5 \
    zlib1g \
    libbz2-1.0 \
    liblz4-1 \
    libzstd1 \
    && rm -rf /var/lib/apt/lists/*


# Copy the RocksDB shared libraries from the build stage
COPY --from=builder /usr/local/lib/librocksdb.so.9.6.1 /usr/local/lib/
COPY --from=builder /usr/local/include/rocksdb /usr/local/include/rocksdb

RUN ln -s /usr/local/lib/librocksdb.so.9.6.1 /usr/local/lib/librocksdb.so.9.6 \
    && ln -s /usr/local/lib/librocksdb.so.9.6 /usr/local/lib/librocksdb.so.9 \
    && ln -s /usr/local/lib/librocksdb.so.9 /usr/local/lib/librocksdb.so

RUN echo "/usr/local/lib" | tee /etc/ld.so.conf.d/rocksdb.conf

RUN ldconfig && ldconfig -p | grep rocksdb

# Copy the binary from the builder stage
COPY --from=builder /app/toyapp .

# Expose any ports (if necessary)
# EXPOSE 8080

# Command to run the application
CMD ["./toyapp"]