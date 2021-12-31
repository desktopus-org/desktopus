#!/bin/bash
# Logger from this post http://www.cubicrace.com/2016/03/log-tracing-mechnism-for-shell-scripts.html

INFO(){
    local msg="$1"
    timeAndDate=$(date)
    echo "[$timeAndDate] [INFO] [${0}] $msg"
}


DEBUG(){
    local msg="$1"
    timeAndDate=$(date)
    echo "[$timeAndDate] [DEBUG] [${0}] $msg"
}

ERROR(){
    local msg="$1"
    timeAndDate=$(date)
    echo "[$timeAndDate] [ERROR]  $msg"
}