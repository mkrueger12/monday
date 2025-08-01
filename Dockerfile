# Build stage
FROM cgr.dev/chainguard/go:latest AS builder

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o monday .

# Runtime stage
FROM node:24-alpine

# Install core utilities and development tools
RUN apk add --no-cache \
    bash \
    git \
    openssh-client \
    curl \
    ca-certificates \
    python3 \
    py3-pip

# Install GitHub CLI (architecture-aware)
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi && \
    if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi && \
    curl -fsSL https://github.com/cli/cli/releases/download/v2.40.1/gh_2.40.1_linux_${ARCH}.tar.gz \
    | tar -xz -C /tmp \
    && mv /tmp/gh_2.40.1_linux_${ARCH}/bin/gh /usr/local/bin/ \
    && rm -rf /tmp/gh_*

RUN apk del curl ca-certificates

# Install OpenAI Codex CLI
RUN npm i -g @openai/codex

# Create app directory
RUN mkdir -p /app

# Copy Monday CLI binary from builder stage
COPY --from=builder /build/monday /usr/local/bin/monday

# Set working directory
WORKDIR /workspace

ENV CODEX_QUIET_MODE=1

EXPOSE 8080

# Set Monday CLI as entrypoint
CMD ["monday", "server"]