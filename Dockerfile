FROM cruizba/ubuntu-dind:19.03.5

ENV HOME=/home/userdocker \
    TERM=xterm-256color \
    DEBIAN_FRONTEND=noninteractive \
    VNC_COL_DEPTH=24 \
    VNC_RESOLUTION=1280x1024 \
    VNC_PW=vncpassword \
    VNC_VIEW_ONLY=false \
    VNC_PORT=5901 \
    NO_VNC_PORT=6901 \
    DISPLAY=:1 \
    USER_DEFAULT_PASSWORD=userpassword

# Installation and supervisor scripts
ADD installation.sh /opt/installation-scripts/
ADD supervisor-scripts /opt/supervisor-scripts/

# XFCE configuration files
ADD ./config-files/ $HOME/

RUN chmod -R +x /opt/supervisor-scripts/ \ 
    && chmod -R +x /opt/installation-scripts/

# Add userdocker user
RUN useradd -m userdocker -s /bin/bash  && echo "userdocker:${USER_DEFAULT_PASSWORD}" | chpasswd && adduser userdocker sudo

RUN cd /opt/installation-scripts && ./installation.sh

# Add supervisor config
COPY supervisor/ /etc/supervisor/conf.d/
COPY startup.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/startup.sh

ENTRYPOINT ["startup.sh"]
CMD ["sh"]