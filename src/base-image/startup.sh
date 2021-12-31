#!/bin/bash
source /opt/bash-utils/logger.sh

function wait_for_process () {
    local max_time_wait=600
    local process_name="$1"
    local waited_sec=0
    while ! pgrep "$process_name" >/dev/null && ((waited_sec < max_time_wait)); do
        INFO "Process $process_name is not running yet. Retrying in 1 seconds"
        INFO "Waited $waited_sec seconds of $max_time_wait seconds"
        sleep 1
        ((waited_sec=waited_sec+1))
        if ((waited_sec >= max_time_wait)); then
            return 1
        fi
    done
    return 0
}

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

INFO "Configuring chrome for docker"

VNC_RES_W=${RESOLUTION%x*}
VNC_RES_H=${RESOLUTION#*x}

echo "CHROMIUM_FLAGS='--no-sandbox --disable-gpu --user-data-dir --window-size=$VNC_RES_W,$VNC_RES_H --window-position=0,0'" > $HOME/.chromium-browser.init

INFO "Creating docker group and adding userdocker to the group"

groupadd docker
usermod -aG docker userdocker

INFO "Adding userdocker to audio group"
usermod -a -G audio userdocker

INFO "Adding userdocker to video and render group for hw acceleration"
usermod -a -G video userdocker
groupadd render
usermod -a -G render userdocker

/usr/bin/supervisord -n