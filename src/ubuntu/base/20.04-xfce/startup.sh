#!/bin/bash
source /opt/bash-utils/logger.sh

# Change password of userdocker
echo "userdocker:${USER_PASSWORD}" | chpasswd

INFO "Configuring VNC Password"

# first entry is control, second is view (if only one is valid for both)
mkdir -p "$HOME/.vnc"
PASSWD_PATH="$HOME/.vnc/passwd"

if [[ -f $PASSWD_PATH ]]; then
    echo -e "\n---------  purging existing VNC password settings  ---------"
    rm -f $PASSWD_PATH
fi
echo "$VNC_PW" | vncpasswd -f >> $PASSWD_PATH
chmod 600 $PASSWD_PATH

INFO "Giving permissions for non root user"
chown -R userdocker /home/userdocker

INFO "Adding userdocker to audio group"
usermod -a -G audio userdocker

INFO "Adding userdocker to video and render group for hw acceleration"
usermod -a -G video userdocker
groupadd render
usermod -a -G render userdocker

/usr/bin/supervisord -n