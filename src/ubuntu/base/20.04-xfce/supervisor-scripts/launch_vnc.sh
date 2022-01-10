#!/bin/bash
### every exit != 0 fails the script
set -e

# Configure tigervnc to start xfce
cat <<EOF > $HOME/.vnc/xstartup
unset SESSION_MANAGER
unset DBUS_SESSION_BUS_ADDRESS
exec startxfce4
EOF

rm "$HOME"/.vnc/*.pid || true
rm "$HOME"/.vnc/*.log || true

# Setting pidfile + command to execute
pidfile="$HOME/.vnc/*:1.pid"
command="vncserver -depth $VNC_COL_DEPTH -geometry $RESOLUTION -localhost no"

# Proxy signals
function kill_app(){
    kill $(cat $pidfile)
    exit 0 # exit okay
}
trap "kill_app" SIGINT SIGTERM

# Launch daemon
$command
sleep 2

# Loop while the pidfile and the process exist
while [ -f $pidfile ] && kill -0 $(cat $pidfile) ; do
    sleep 0.5
done
exit 1000 # exit unexpected