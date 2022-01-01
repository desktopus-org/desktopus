#!/bin/bash

export TIMEZONE=$(cat /etc/timezone)
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

function help_msg()
{
   # Display Help
   echo "You can execute..."
   echo
   echo "./run_workspace.sh --basic (No audio)"
   echo "./run_workspace.sh --basic-privileged (No audio) (If you have docker installed and want to use Docker in Docker)"
   echo "./run_workspace.sh --audio (With audio)"
   echo "./run_workspace.sh --audio-privileged (With audio) (If you have docker installed and want to use Docker in Docker)"
   echo
}

function run_basic() {
    local EXTRA_ARGS=${1:-}
    docker run --rm -it $(if [ -n "$EXTRA_ARGS" ]; then echo "$EXTRA_ARGS"; fi) \
        -p 5901:5901 -p 6901:6901 \
        --shm-size=256m \
        --env RESOLUTION="${RESOLUTION:-1920x1080}" \
        --env USER_PASSWORD="${USER_PASSWORD:-userpassword}" \
        --env VNC_PW="${VNC_PW:-vncpassword}" \
        --env TZ="${TIMEZONE}" \
        --volume "$(pwd)"/shared-home:/home/desktopus/shared-home \
        __workspace_name__
}

function run_audio() {
    local EXTRA_ARGS=${1:-}
    # Create pulseaudio socket
    pactl load-module module-native-protocol-unix socket=/tmp/pulseaudio.socket

    # Create /tmp/pulseaudio.client.conf for pulseaudio clients
    create_pulse_audio_conf

    # Docker run
    docker run --rm -it $(if [ -n "$EXTRA_ARGS" ]; then echo "$EXTRA_ARGS"; fi) \
    -p 5901:5901 -p 6901:6901 \
    --shm-size=256m \
    --env PULSE_SERVER=unix:/tmp/pulseaudio.socket \
    --env PULSE_COOKIE=/tmp/pulseaudio.cookie \
    --env RESOLUTION="${RESOLUTION:-1920x1080}" \
    --env USER_PASSWORD="${USER_PASSWORD:-userpassword}" \
    --env VNC_PW="${VNC_PW:-vncpassword}" \
    --env TZ="${TIMEZONE}" \
    --volume /tmp/pulseaudio.socket:/tmp/pulseaudio.socket \
    --volume /tmp/pulseaudio.client.conf:/etc/pulse/client.conf \
    --volume "$(pwd)"/shared-home:/home/desktopus/shared-home \
    __workspace_name__
}

if [ "$#" -gt 0 ]; then
    if [ "$1" = "--basic-privileged" ]; then
        run_basic --privileged
    elif [ "$1" = "--basic" ]; then
        run_basic
    elif [ "$1" = "--audio-privileged" ]; then
        run_audio --privileged
    elif [ "$1" = "--audio" ]; then
        run_audio
    elif [ "$1" == "--help" ]; then
        help_msg
    else
        help_msg
    fi
else 
    help_msg
fi