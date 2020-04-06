#!/usr/bin/env bash
##################
# cluster-dev installation script
# example command ./cluster-dev install
#


# sample reading values from customre with pre-defined values
read -p "Please enter the name of your infrastructure repository [infrastructure]: " name
name=${name:-infrastructure}
echo $name

read -e -p "Please enter the name of your infrastructure repository: " -i "infrastructure" NAME
echo $NAME


# sample menu supporting arrows
#
declare -a menu=("Github" "Bitbucket" "Gitlab")
cur=0

draw_menu() {
    for i in "${menu[@]}"; do
        if [[ ${menu[$cur]} == $i ]]; then
            tput setaf 2; echo " > $i"; tput sgr0
        else
            echo "   $i";
        fi
    done
}

clear_menu()  {
    for i in "${menu[@]}"; do tput cuu1; done
    tput ed
}

# Draw initial Menu
draw_menu
while read -sN1 key; do # 1 char (not delimiter), silent
    # Check for enter/space
    if [[ "$key" == "" ]]; then break; fi

    # catch multi-char special key sequences
    read -sN1 -t 0.0001 k1; read -sN1 -t 0.0001 k2; read -sN1 -t 0.0001 k3
    key+=${k1}${k2}${k3}

    case "$key" in
        # cursor up, left: previous item
        i|j|$'\e[A'|$'\e0A'|$'\e[D'|$'\e0D') ((cur > 0)) && ((cur--));;
        # cursor down, right: next item
        k|l|$'\e[B'|$'\e0B'|$'\e[C'|$'\e0C') ((cur < ${#menu[@]}-1)) && ((cur++));;
        # home: first item
        $'\e[1~'|$'\e0H'|$'\e[H')  cur=0;;
        # end: last item
        $'\e[4~'|$'\e0F'|$'\e[F') ((cur=${#menu[@]}-1));;
        # q, carriage return: quit
        q|''|$'\e')echo "Aborted." && exit;;
    esac
    # Redraw menu
    clear_menu
    draw_menu
done

echo "Selected item $cur: ${menu[$cur]}";
