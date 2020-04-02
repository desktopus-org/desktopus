# Ubuntu XFCE docker (With Docker in Docker, audio and video acceleration)

A reproducible ubuntu desktop in docker with docker!

### OpenPorts:

- 5901 - VNC
- 6901 - noVNC

## Basic Run:

It supports:
- Docker in Docker

```
./run_ubuntu.sh --basic
```

## Run with audio(pulseaudio) and hardware acceleration:

It suports:
- Audio
- Video HW acceleration
- Docker in Docker

```
./run_ubuntu.sh --audio-video
```

## Credentials

Ubuntu user password: userpassword

VCN Password: vncpassword
