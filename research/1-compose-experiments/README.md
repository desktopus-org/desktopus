# Docker examples with TigerVNC

This example let you run a desktop environment in a container which automatically opens TigerVNC server and connects to it.

## Build

```
docker build . -f Dockerfile.base -t desktopus/core-ubuntu-jammy:latest
docker compose build
```

## Run (with cpu)

```
docker compose up
```

## Run (with gpu using DRI3)

```
docker compose -f docker-compose.gpu.yml up
```

## Run (with gpu using VirtualGL)

```
docker compose -f docker-compose.gpu.egl.yml up
```
