# Multi-Platform Build Guide

This guide explains how to build the inventory service for multiple platforms (amd64, arm64, etc.).

## Prerequisites

### For Docker
```bash
# Install Docker Buildx (usually included in Docker Desktop)
docker buildx version

# Create a new builder instance (one-time setup)
make buildx-setup
```

### For Podman
```bash
# Podman supports multi-platform builds natively
podman --version
```

## Build Commands

### 1. Build for Current Platform
```bash
# Using Makefile
make build

# Or directly
podman build -t inventory-service:latest .
```

### 2. Build for Specific Platform

#### AMD64 (x86_64)
```bash
make build-amd64

# Or directly
podman build --platform linux/amd64 -t inventory-service:latest-amd64 .
```

#### ARM64 (Apple Silicon, AWS Graviton)
```bash
make build-arm64

# Or directly
podman build --platform linux/arm64 -t inventory-service:latest-arm64 .
```

### 3. Build for Multiple Platforms

#### Using Docker Buildx
```bash
# Build for both amd64 and arm64
make build-multi

# Or directly
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t inventory-service:latest \
  .
```

#### Build and Push to Registry
```bash
# Update registry name in Makefile first
make build-push-multi

# Or directly
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t your-registry.com/inventory-service:latest \
  --push \
  .
```

## Dockerfile Features

### Multi-Stage Build
- **Builder stage**: Compiles Go binary for target platform
- **Runtime stage**: Minimal Alpine image with only necessary dependencies

### Cross-Compilation Support
```dockerfile
ARG TARGETOS    # Target OS (linux, darwin, windows)
ARG TARGETARCH  # Target architecture (amd64, arm64, arm)
```

### Security Features
- Runs as non-root user (appuser:1000)
- Minimal attack surface (Alpine base)
- Static binary with no external dependencies

### Optimizations
- Layer caching for go.mod/go.sum
- Stripped binary (`-ldflags='-w -s'`)
- Static linking (`-extldflags "-static"`)

## Platform-Specific Notes

### AMD64 (x86_64)
- Most common platform
- Used in most cloud providers
- Best compatibility

### ARM64 (aarch64)
- Apple Silicon (M1, M2, M3)
- AWS Graviton instances
- Raspberry Pi 4+
- Better power efficiency

### ARM (32-bit)
- Raspberry Pi 3 and older
- IoT devices
- Limited support

## Testing Multi-Platform Images

### Inspect Image Platforms
```bash
# Using Docker
docker buildx imagetools inspect inventory-service:latest

# Using Podman
podman manifest inspect inventory-service:latest
```

### Run on Different Platforms
```bash
# Force run on specific platform
podman run --platform linux/amd64 inventory-service:latest
podman run --platform linux/arm64 inventory-service:latest
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Build Multi-Platform

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: your-registry/inventory-service:latest
```

### GitLab CI Example
```yaml
build:
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker buildx create --use
    - docker buildx build --platform linux/amd64,linux/arm64 -t inventory-service:latest .
```

## Troubleshooting

### QEMU Not Found
```bash
# Install QEMU for cross-platform emulation
docker run --privileged --rm tonistiigi/binfmt --install all
```

### Build Fails on ARM64
```bash
# Ensure you have enough memory (4GB+ recommended)
# Check Docker/Podman resource limits
```

### Slow Cross-Platform Builds
- Cross-compilation is slower due to QEMU emulation
- Consider using native builders for each platform
- Use build cache to speed up subsequent builds

## Performance Comparison

| Platform | Build Time | Image Size | Runtime Performance |
|----------|-----------|------------|---------------------|
| AMD64    | ~2-3 min  | ~15 MB     | Baseline            |
| ARM64    | ~2-3 min  | ~15 MB     | 10-20% better efficiency |
| ARM      | ~3-5 min  | ~14 MB     | 30-40% slower       |

## Best Practices

1. **Use Multi-Stage Builds**: Reduces final image size
2. **Cache Dependencies**: Copy go.mod/go.sum first
3. **Static Linking**: Ensures portability across platforms
4. **Security**: Run as non-root user
5. **Minimal Base**: Use Alpine for smaller images
6. **Version Tags**: Tag images with version numbers
7. **Platform Tags**: Tag platform-specific images clearly

## Example Deployment

### Docker Compose with Platform Selection
```yaml
version: '3.8'

services:
  inventory-service:
    image: inventory-service:latest
    platform: linux/amd64  # or linux/arm64
    ports:
      - "13740:13740"
    environment:
      - DB_SOURCE=${DB_SOURCE}
```

### Kubernetes with Node Affinity
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: inventory-service
spec:
  template:
    spec:
      containers:
      - name: inventory-service
        image: inventory-service:latest
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: kubernetes.io/arch
                operator: In
                values:
                - amd64
                - arm64
```

## Resources

- [Docker Buildx Documentation](https://docs.docker.com/buildx/working-with-buildx/)
- [Podman Multi-Arch](https://www.redhat.com/sysadmin/multi-architecture-images-podman)
- [Go Cross-Compilation](https://go.dev/doc/install/source#environment)
