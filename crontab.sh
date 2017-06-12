#!/usr/bin/env bash
GetList () {
    crontab -l > mycron
}

Add () {
    # crontab -l > mycron
    crontab mycron
    rm mycron
}

option="${1}"
case ${option} in
    -l)
        GetList
    ;;
    -c)
        Add
    ;;
    -h)
        echo "`basename ${0}`:usage: [-c create] | [-d delete] [crontab_syntax]"
        exit 1
    ;;
esac