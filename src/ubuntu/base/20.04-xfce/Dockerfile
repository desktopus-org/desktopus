FROM ubuntu:20.04

# Installation for docker in docker
ARG NOVNC_VERSION=1.3.0

ENV HOME=/home/desktopus \
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

# Base installation
COPY base_installation.sh /opt/base-installation-scripts/
ADD supervisor-scripts /opt/supervisor-scripts/

# XFCE configuration files
ADD ./config-files/ $HOME/

RUN chmod -R +x /opt/supervisor-scripts/ \ 
    && chmod -R +x /opt/base-installation-scripts/

# Add desktopus user
RUN useradd -m desktopus -s /bin/bash && adduser desktopus sudo

RUN cd /opt/base-installation-scripts/ && ./base_installation.sh
ADD modules-installation /opt/modules-installation/
RUN chmod -R +x /opt/modules-installation/

RUN cd /opt/modules-installation/ && \
    count=`ls -1 *.sh 2>/dev/null | wc -l` && \
    if [ "${count}" != 0 ]; then \
        for MODULE_SCRIPT in *.sh; do bash "$MODULE_SCRIPT"; done \
    fi

# Add supervisor config
COPY supervisor/ /etc/supervisor/conf.d/
COPY startup.sh /usr/local/bin/
COPY logger.sh /opt/bash-utils/logger.sh
RUN chmod +x /usr/local/bin/startup.sh
RUN chown desktopus:desktopus /home/desktopus

ENTRYPOINT ["startup.sh"]