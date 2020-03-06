#!/bin/bash

function create_pulse_audio_conf () {

cat <<EOF >/tmp/pulseaudio.client.conf
default-server = unix:/tmp/pulseaudio.socket
# Prevent a server running in the container
autospawn = no
daemon-binary = /bin/true
# Prevent the use of shared memory
enable-shm = false
EOF

}

if [ "$#" -gt 0 ]; then
    if [ "$1" = "--basic" ]; then
        docker run -it --privileged \
        -p 5901:5901 -p 6901:6901 \
        --shm-size=256m \
        cruizba/ubuntu-xfce-docker
    fi
    if [ "$1" = "--audio-video" ]; then

        # Create pulseaudio socket
        pactl load-module module-native-protocol-unix socket=/tmp/pulseaudio.socket

        # Create /tmp/pulseaudio.client.conf for pulseaudio clients
        create_pulse_audio_conf

        # Docker run
        docker run -it --privileged \
        -p 5901:5901 -p 6901:6901 \
        --shm-size=256m \
        --env PULSE_SERVER=unix:/tmp/pulseaudio.socket \
        --env PULSE_COOKIE=/tmp/pulseaudio.cookie \
        --env VNC_RESOLUTION=1440x900 \
        --volume /tmp/pulseaudio.socket:/tmp/pulseaudio.socket \
        --volume /tmp/pulseaudio.client.conf:/etc/pulse/client.conf \
        --device /dev/dri \
        --device /dev/vga_arbiter \
        --device /dev/video0 \
        --device /dev/snd:/dev/snd \
        cruizba/ubuntu-xfce-docker
    fi
fi