FROM ubuntu:20.04

# Installation for docker in docker
ENV DOCKER_CHANNEL=stable
ARG DOCKER_VERSION=20.10.12
ARG DOCKER_COMPOSE_VERSION=1.29.2
ARG NOVNC_VERSION=1.3.0

ENV HOME=/home/userdocker \
    TERM=xterm-256color \
    DEBIAN_FRONTEND=noninteractive \
    VNC_COL_DEPTH=24 \
    RESOLUTION=1920x1080 \
    VNC_PW=vncpassword \
    VNC_VIEW_ONLY=false \
    VNC_PORT=5901 \
    NO_VNC_PORT=6901 \
    DISPLAY=:1 \
    USER_PASSWORD=userpassword

# Installation and supervisor scripts
ADD installation.sh /opt/installation-scripts/
ADD supervisor-scripts /opt/supervisor-scripts/

# XFCE configuration files
ADD ./config-files/ $HOME/

RUN chmod -R +x /opt/supervisor-scripts/ \ 
    && chmod -R +x /opt/installation-scripts/

# Add userdocker user
RUN useradd -m userdocker -s /bin/bash && adduser userdocker sudo

RUN cd /opt/installation-scripts && ./installation.sh

# Add supervisor config
COPY supervisor/ /etc/supervisor/conf.d/
COPY startup.sh /usr/local/bin/
COPY logger.sh /opt/bash-utils/logger.sh
RUN chmod +x /usr/local/bin/startup.sh

ENTRYPOINT ["startup.sh"]