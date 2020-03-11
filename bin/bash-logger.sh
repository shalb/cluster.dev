#!/bin/bash
#--------------------------------------------------------------------------------------------------
# Bash Logger
# Copyright (c) Dean Rather
# Licensed under the MIT license
# http://github.com/deanrather/bash-logger
#--------------------------------------------------------------------------------------------------

#--------------------------------------------------------------------------------------------------
# Configurables

# Minimum log level to show. Default to DEBUG
readonly LOG_LVL=${VERBOSE_LVL-"DEBUG"}

readonly LOGFILE=/dev/null                    # Writes logs only to stdout

readonly LOG_FORMAT="%DATE PID:%PID func:'%FUNC_NAME' %LEVEL - %MESSAGE" # Eg: 2020-03-10 16:18:31 UTC+02 PID:29871 Run func:'main' DEBUG - Example Debug log
readonly LOG_DATE_FORMAT='+%F %T UTC%:::z'                               # Eg: 2020-03-10 16:18:31 UTC+02

readonly LOG_FORMAT_SIMPLE="%DATE - %MESSAGE" # Eg: 16:29:34 - Example Info log
readonly LOG_DATE_FORMAT_SIMPLE='+%T'         # Eg: 16:29:34

readonly LOG_COLOR_DEBUG="\e[38;5;247m"       # Grey
readonly LOG_COLOR_INFO="\e[94m"              # Blue
readonly LOG_COLOR_NOTICE="\033[1;32m"        # Default: Bold Green
readonly LOG_COLOR_WARNING="\033[1;33m"       # Bold Yellow
readonly LOG_COLOR_ERROR="\033[1;31m"         # Bold Red
readonly LOG_COLOR_CRITICAL="\e[1;48;5;88m"   # Bold White Text, Dark Red Background
readonly LOG_COLOR_ALERT="\e[1;41m"           # Bold White Text, Red Background
readonly LOG_COLOR_EMERGENCY="\e[1;38;5;52m\e[48;5;196m" # Bold Dark Red Text, Light Red Background
readonly RESET_COLOR="\033[0m"

#--------------------------------------------------------------------------------------------------
# Individual Log Functions
# These can be overwritten to provide custom behavior for different log levels

DEBUG()     { LOG_HANDLER_DEFAULT "$@"; }
INFO()      { LOG_HANDLER_DEFAULT "$@" "$LOG_FORMAT_SIMPLE" "$LOG_DATE_FORMAT_SIMPLE"; }
NOTICE()    { LOG_HANDLER_DEFAULT "$@" "$LOG_FORMAT_SIMPLE" "$LOG_DATE_FORMAT_SIMPLE"; }
WARNING()   { LOG_HANDLER_DEFAULT "$@" "$LOG_FORMAT_SIMPLE" "$LOG_DATE_FORMAT_SIMPLE"; }
# Print empty lines before and on the end of Error+ logs
ERROR()     { echo; LOG_HANDLER_DEFAULT "$@"; echo; exit 1;}
CRITICAL()  { echo; LOG_HANDLER_DEFAULT "$@"; echo; exit 1;}
ALERT()     { echo; LOG_HANDLER_DEFAULT "$@"; echo; exit 1;}
EMERGENCY() { echo; LOG_HANDLER_DEFAULT "$@"; echo; exit 1;}

#--------------------------------------------------------------------------------------------------
# Helper Functions

# Outputs a log formatted using the LOG_FORMAT and DATE_FORMAT configurables
# Usage: FORMAT_LOG <log level> <log message>
# Eg: FORMAT_LOG CRITICAL "My critical log"
FORMAT_LOG() {
    local level="$1"
    local log="$2"
    local formatted_log="${3:-"$LOG_FORMAT"}"
    local date_format="${4:-"$LOG_DATE_FORMAT"}"
    local func_name="$5"
    local pid=$$
    local date
    date="$(date "$date_format")"

    formatted_log="${formatted_log/'%MESSAGE'/$log}"
    formatted_log="${formatted_log/'%LEVEL'/$level}"
    formatted_log="${formatted_log/'%PID'/$pid}"
    formatted_log="${formatted_log/'%DATE'/$date}"
    formatted_log="${formatted_log/'%FUNC_NAME'/$func_name}"
    echo "$formatted_log"
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
    # $1 - message
    local log_format=${2:-"$LOG_FORMAT"}
    local log_date_format=${3:-"$LOG_DATE_FORMAT"}

    local func_name=${FUNCNAME[2]}
    local lvl=${FUNCNAME[1]} # Log level get from function name

    # Disable logging by LOG_LVL
    case $LOG_LVL in
        DEBUG)
            ;;
        INFO)
            if  [ "$lvl" == DEBUG    ]; then return; fi
            ;;
        NOTICE)
            if  [ "$lvl" == DEBUG    ] || \
                [ "$lvl" == INFO     ]; then return; fi
            ;;
        WARNING)
            if  [ "$lvl" == DEBUG    ] || \
                [ "$lvl" == INFO     ] || \
                [ "$lvl" == NOTICE   ]; then return; fi
            ;;
        ERROR)
            if  [ "$lvl" == DEBUG    ] || \
                [ "$lvl" == INFO     ] || \
                [ "$lvl" == NOTICE   ] || \
                [ "$lvl" == WARNING  ]; then return; fi
            ;;
        CRITICAL)
            if  [ "$lvl" == DEBUG    ] || \
                [ "$lvl" == INFO     ] || \
                [ "$lvl" == NOTICE   ] || \
                [ "$lvl" == WARNING  ] || \
                [ "$lvl" == ERROR    ]; then return; fi
            ;;
        ALERT)
            if  [ "$lvl" == DEBUG    ] || \
                [ "$lvl" == INFO     ] || \
                [ "$lvl" == NOTICE   ] || \
                [ "$lvl" == WARNING  ] || \
                [ "$lvl" == ERROR    ] || \
                [ "$lvl" == CRITICAL ]; then return; fi
            ;;
        EMERGENCY)
            if  [ "$lvl" == DEBUG    ] || \
                [ "$lvl" == INFO     ] || \
                [ "$lvl" == NOTICE   ] || \
                [ "$lvl" == WARNING  ] || \
                [ "$lvl" == ERROR    ] || \
                [ "$lvl" == CRITICAL ] || \
                [ "$lvl" == ALERT    ]; then return; fi
            ;;
        *)
            ;;
    esac

    local formatted_log
    formatted_log="$(FORMAT_LOG "$lvl" "$@" "$log_format" "$log_date_format" "$func_name")"

    LOG_HANDLER_COLORTERM "$lvl" "$formatted_log"
    LOG_HANDLER_LOGFILE "$lvl" "$formatted_log"
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
    echo -e "$log"
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
