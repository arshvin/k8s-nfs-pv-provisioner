C_RED=$(tput setaf 1)
C_GREEN=$(tput setaf 2)
C_WHITE=$(tput setaf 7)
C_BLUE=$(tput setaf 4)
C_RESET=$(tput sgr0)

# @see https://unix.stackexchange.com/a/346093
export PS1="\[${C_BLUE}\][\[${C_WHITE}\]\u\[${C_BLUE}\]]\[${C_RESET}\] \w \[${C_GREEN}\]>\[${C_RESET}\] "
