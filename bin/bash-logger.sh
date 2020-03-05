#!/bin/bash
#--------------------------------------------------------------------------------------------------
# Bash Logger
# Copyright (c) Dean Rather
# Licensed under the MIT license
# http://github.com/deanrather/bash-logger
#--------------------------------------------------------------------------------------------------

#--------------------------------------------------------------------------------------------------
# Configurables

export LOGFILE=/dev/null                    # Writes logs only to stdout
export LOG_FORMAT="%DATE PID:%PID %LEVEL - %MESSAGE" # Eg: 2020-03-04 16:53:22 UTC+02 DEBUG: Example Debug log
export LOG_DATE_FORMAT='+%F %T UTC%:::z'    # Eg: 2020-03-04 16:30:01 UTC+02
export LOG_COLOR_DEBUG="\e[38;5;247m"       # Grey
export LOG_COLOR_INFO="\e[94m"              # Blue
export LOG_COLOR_NOTICE="\033[1;32m"        # Default: Bold Green
export LOG_COLOR_WARNING="\033[1;33m"       # Bold Yellow
export LOG_COLOR_ERROR="\033[1;31m"         # Bold Red
export LOG_COLOR_CRITICAL="\e[1;48;5;88m"   # Bold White Text, Dark Red Background
export LOG_COLOR_ALERT="\e[1;41m"           # Bold White Text, Red Background
export LOG_COLOR_EMERGENCY="\e[1;38;5;52m\e[48;5;196m" # Bold Dark Red Text, Light Red Background
export RESET_COLOR="\033[0m"

#--------------------------------------------------------------------------------------------------
# Individual Log Functions
# These can be overwritten to provide custom behavior for different log levels

DEBUG()     { LOG_HANDLER_DEFAULT "${FUNCNAME[@]}" "$@"; }
INFO()      { LOG_HANDLER_DEFAULT "${FUNCNAME[@]}" "$@"; }
NOTICE()    { LOG_HANDLER_DEFAULT "${FUNCNAME[@]}" "$@"; }
WARNING()   { LOG_HANDLER_DEFAULT "${FUNCNAME[@]}" "$@"; }
# Print empty lines before and on the end of Error+ logs
ERROR()     { echo; LOG_HANDLER_DEFAULT "${FUNCNAME[@]}" "$@"; echo; exit 1;}
CRITICAL()  { echo; LOG_HANDLER_DEFAULT "${FUNCNAME[@]}" "$@"; echo; exit 1;}
ALERT()     { echo; LOG_HANDLER_DEFAULT "${FUNCNAME[@]}" "$@"; echo; exit 1;}
EMERGENCY() { echo; LOG_HANDLER_DEFAULT "${FUNCNAME[@]}" "$@"; echo; exit 1;}

#--------------------------------------------------------------------------------------------------
# Helper Functions

# Outputs a log formatted using the LOG_FORMAT and DATE_FORMAT configurables
# Usage: FORMAT_LOG <log level> <log message>
# Eg: FORMAT_LOG CRITICAL "My critical log"
FORMAT_LOG() {
    local level="$1"
    local log="$2"
    local pid=$$
    local date
    date="$(date "$LOG_DATE_FORMAT")"

    local formatted_log="$LOG_FORMAT"
    formatted_log="${formatted_log/'%MESSAGE'/$log}"
    formatted_log="${formatted_log/'%LEVEL'/$level}"
    formatted_log="${formatted_log/'%PID'/$pid}"
    formatted_log="${formatted_log/'%DATE'/$date}"
    # shellcheck disable=SC2028
    echo "$formatted_log\n"
}

# Calls one of the individual log functions
# Usage: LOG <log level> <log message>
# Eg: LOG INFO "My info log"
LOG() {
    local level="$1"
    local log="$2"
    local log_function_name="${!level}"
    $log_function_name "$log"
}

#--------------------------------------------------------------------------------------------------
# Log Handlers

# All log levels call this handler (by default...), so this is a great place to put any standard
# logging behavior
# Usage: LOG_HANDLER_DEFAULT <log level> <log message>
# Eg: LOG_HANDLER_DEFAULT DEBUG "My debug log"
LOG_HANDLER_DEFAULT() {
    # $1 - level
    # $2 - message
    local formatted_log

    formatted_log="$(FORMAT_LOG "$@")"
    LOG_HANDLER_COLORTERM "$1" "$formatted_log"
    LOG_HANDLER_LOGFILE "$1" "$formatted_log"
}




# Outputs a log to the stdout, colorized using the LOG_COLOR configurables
# Usage: LOG_HANDLER_COLORTERM <log level> <log message>
# Eg: LOG_HANDLER_COLORTERM CRITICAL "My critical log"
LOG_HANDLER_COLORTERM() {
    local level="$1"
    local log="$2"
    local color_variable="LOG_COLOR_$level"
    local color="${!color_variable}"
    log="$color$log$RESET_COLOR"
    echo -en "$log"
}

# Appends a log to the configured logfile
# Usage: LOG_HANDLER_LOGFILE <log level> <log message>
# Eg: LOG_HANDLER_LOGFILE NOTICE "My critical log"
LOG_HANDLER_LOGFILE() {
    local level="$1"
    local log="$2"
    local log_path

    log_path="$(dirname "$LOGFILE")"
    [ -d "$log_path" ] || mkdir -p "$log_path"
    echo "$log" >> "$LOGFILE"
}
