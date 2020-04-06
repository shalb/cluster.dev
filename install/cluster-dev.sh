#!/usr/bin/env bash
##################
# cluster-dev installation script
# example command ./cluster-dev install
#

#######################################
# Draw menu and user select one of items
# Globals:
#   FUNC_RESULT
# Arguments:
#   menu - array. menu options
#   cur - number. cursor position
# Outputs:
#   Draw menu, Mutate FUNC_RESULT to ${menu[$cur]}
#######################################
function interactive_select {
    local menu=("$@")
    local cur=0

    # Check is `cur` set
    if [ -n "${menu[-1]}" ] && [ "${menu[-1]}" -eq "${menu[-1]}" ] 2>/dev/null; then
        cur=${menu[-1]}
        menu=("${menu[@]:0:${#menu[@]}-1}")
    fi

    # Draw menu
    for i in "${menu[@]}"; do
        if [[ ${menu[$cur]} == "$i" ]]; then
            tput setaf 2; echo " > $i"; tput sgr0
        else
            echo "   $i";
        fi
    done

    while read -r -sN1 key; do # 1 char (not delimiter), silent
        # Check for enter/space
        if [[ "$key" == "" ]]; then break; fi

        # catch multi-char special key sequences
        read -r -sN1 -t 0.0001 k1; read -r -sN1 -t 0.0001 k2; read -r -sN1 -t 0.0001 k3
        key+=${k1}${k2}${k3}

        case "$key" in
            # cursor up, left: previous item
            i|j|$'\e[A'|$'\e0A'|$'\e[D'|$'\e0D') ((cur > 0)) && ((cur--));;
            # cursor down, right: next item
            k|l|$'\e[B'|$'\e0B'|$'\e[C'|$'\e0C') ((cur < ${#menu[@]}-1)) && ((cur++));;
            # home: first item
            $'\e[1~'|$'\e0H'|$'\e[H') cur=0;;
            # end: last item
            $'\e[4~'|$'\e0F'|$'\e[F') ((cur=${#menu[@]}-1));;
            # carriage return: select
            $'\n') echo && FUNC_RESULT="${menu[$cur]}" && return;;
        esac
        # Redraw menu
        clear_menu "${menu[@]}"
        interactive_select "${menu[@]}" $cur && return # exit from recursive levels
    done
}

#######################################
# Remove menu from terminal
# Globals:
#   None
# Arguments:
#   menu - array. menu options
# Outputs:
#   Cleanup menu
#######################################
function clear_menu {
    local menu=("$@")

    for i in "${menu[@]}"; do tput cuu1; done
    tput ed
}

#######################################
# Check if current directory is a Git repository
# Globals:
#   FUNC_RESULT
# Arguments:
#   None
# Outputs:
#   Mutate FUNC_RESULT to boolean
#######################################
function is_current_dir_is_git_repo {
    [ -d .git ] || git rev-parse --git-dir > /dev/null 2>&1
    # shellcheck disable=SC2181
    if [ $? = 0 ]; then
        FUNC_RESULT=true
        return
    fi
    FUNC_RESULT=false
}

#######################################
# Set variables from CLI params
# Globals:
#   None
# Arguments:
#   $@ - key=value format
# Outputs:
#   Declare variables
#######################################
function params {
    # Exit if no subcommand provided
    [ "$1" ] || { echo "No arguments provided. Run './cluster-dev -h' for help"; exit 0; }
    # Skip subcommand
    [ "$1" = 'install' ] && shift
    # Return if no CLI options provided
    [ "$1" ] || return
    local params=("$@")
    local name
    local value

    echo "Provided params:"
    for i in "${params[@]}"; do
        echo "    $i"
        name=$(echo "$i" | awk -F'=' '{print $1}')
        value=$(echo "$i" | awk -F'=' '{print $2}')

        declare -n __var__="$name"
        # shellcheck disable=SC2034
        __var__="$value"
    done
    echo
}

#######################################
# Show help message is '-h' provided and exit
# Globals:
#   None
# Arguments:
#   $@ - -h
# Outputs:
#   Help message
#######################################
function show_help {
    while getopts ":h" opt; do
        case ${opt} in
            h )
            echo "Usage:"
            echo "    cluster-dev -h                      Display this help message."
            echo "    cluster-dev install <params>        Install <with params>."
            exit 0
            ;;
        \? )
            echo "Invalid Option: -$OPTARG" 1>&2
            exit 1
            ;;
        esac
    done
    shift $((OPTIND -1))

    subcommand=$1; shift  # Remove 'pip' from the argument list
    case "$subcommand" in
        # Parse options to the install sub command
        install)
            while getopts ":h" opt; do
                case ${opt} in
                    h )
                    echo "Usage:"
                    echo "    cluster-dev install -h              Display this help message."
                    echo "    cluster-dev install <params>        Install <with params>."
                    echo "      Available params:"
                    echo "      git_provider={Github,Bitbucket,Gitlab}"
                    exit 0
                    ;;
                \? )
                    echo "Invalid Option: -$OPTARG" 1>&2
                    exit 1
                    ;;
                esac
            done
            shift $((OPTIND -1))
        ;;
    esac
}



#######################################################################
#                               M A I N                               #
#######################################################################


# show help if '-h' provided
show_help "$@"

# Set CLI params
git_provider=''
# Get variables value from CLI
params "$@"


declare -a git_providers=(
    "Github"
    "Bitbucket"
    "Gitlab"
)

# Check params
if [[ ! ${git_providers[*]} =~ ${git_provider} ]]; then
    # whatever you want to do when arr doesn't contain value
    echo -e "Parameter 'git_provider=$git_provider' has invalid value.\nAccepted values is: ${git_providers[*]}"
    exit 1
fi


FUNC_RESULT='' # variable used as return from functions


echo -e "Hi, we gonna create an infrastructure for you.\n"

is_current_dir_is_git_repo
if [ "$FUNC_RESULT" = false ]; then
    echo "As this is a GitOps approach we need to start with the git repo"

    create_repo="We could create a repo for you"
    user_turn="You can create or clone by your own, and then run tool there"
    declare -a options=(
        "$create_repo"
        "$user_turn"
    )
    interactive_select "${options[@]}"

    if [ "$FUNC_RESULT" = "$user_turn" ]; then
        echo "OK. See you soon!"
        exit 0;
    fi

    if [ "$FUNC_RESULT" = "$create_repo" ]; then
        # Select github provider
        echo "Please, select your Git hosting:"
        if [ -z "$git_provider" ]; then
            interactive_select "${git_providers[@]}"
        else
            echo "$git_provider"
            FUNC_RESULT="$git_provider"
        fi

        # TODO: check credentials
        exit 0

    fi

fi

echo "Inside git repo, use it."




# # sample reading values from customer with pre-defined values
# read -r -p "Please enter the name of your infrastructure repository [infrastructure]: " name
# name=${name:-infrastructure}
# echo "$name"

# read -r -e -p "Please enter the name of your infrastructure repository: " -i "infrastructure" NAME
# echo "$NAME"
