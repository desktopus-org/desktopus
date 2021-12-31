# Set path
PATH=$PATH

export PS1="\[$(tput bold)\]\[\033[38;5;46m\]\u\[$(tput sgr0)\]\[\033[38;5;15m\]\[$(tput bold)\]:\[$(tput sgr0)\]\[\033[38;5;69m\]\[$(tput bold)\]\W\[$(tput sgr0)\]\[\033[38;5;15m\]\[$(tput bold)\]\\$\[$(tput sgr0)\] \[$(tput sgr0)\]"

alias ls='ls --color=auto'
alias grep='grep --color=auto'
alias fgrep='fgrep --color=auto'
alias egrep='egrep --color=auto'