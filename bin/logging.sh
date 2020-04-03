#!/bin/bash
#
# PSR-3 compliant logging
# See docs at docs/bash-logging.md
#
# Fork of http://github.com/deanrather/bash-logger
#

# Configurables

# Minimum log level to show. Default to INFO
LOG_LVL=${VERBOSE_LVL:-"INFO"}

LOGFILE=/dev/null                    # Writes logs only to stdout

LOG_FORMAT="%DATE PID:%PID func trace: '%FUNC_TRACE' %LEVEL - %MESSAGE" # Eg: 2020-03-10 16:18:31 UTC+02 PID:29871 Run func:'main' DEBUG - Example Debug log
LOG_DATE_FORMAT='+%F %T UTC%:::z'                               # Eg: 2020-03-10 16:18:31 UTC+02

LOG_FORMAT_SIMPLE="%DATE - %MESSAGE" # Eg: 16:29:34 - Example Info log
LOG_DATE_FORMAT_SIMPLE='+%T'         # Eg: 16:29:34

# shellcheck disable=SC2034
LOG_COLOR_DEBUG="\e[38;5;247m"       # Grey
# shellcheck disable=SC2034
LOG_COLOR_INFO="\e[94m"              # Blue
# shellcheck disable=SC2034
LOG_COLOR_NOTICE="\033[1;32m"        # Default: Bold Green
# shellcheck disable=SC2034
LOG_COLOR_WARNING="\033[1;33m"       # Bold Yellow
# shellcheck disable=SC2034
LOG_COLOR_ERROR="\033[1;31m"         # Bold Red
# shellcheck disable=SC2034
LOG_COLOR_CRITICAL="\e[1;48;5;88m"   # Bold White Text, Dark Red Background
# shellcheck disable=SC2034
LOG_COLOR_ALERT="\e[1;41m"           # Bold White Text, Red Background
# shellcheck disable=SC2034
LOG_COLOR_EMERGENCY="\e[1;38;5;52m\e[48;5;196m" # Bold Dark Red Text, Light Red Background
RESET_COLOR="\033[0m"

#######################################
# Individual Log Functions
# These can be overwritten to provide custom behavior for different log levels
# Have no required arguments
# Globals:
#   None
# Arguments:
#   See LOG_HANDLER_DEFAULT()
#######################################
function DEBUG     { LOG_HANDLER_DEFAULT "$@"; }
function INFO      { LOG_HANDLER_DEFAULT "$@" "$LOG_FORMAT_SIMPLE" "$LOG_DATE_FORMAT_SIMPLE"; }
function NOTICE    { LOG_HANDLER_DEFAULT "$@" "$LOG_FORMAT_SIMPLE" "$LOG_DATE_FORMAT_SIMPLE"; }
function WARNING   { LOG_HANDLER_DEFAULT "$@" "$LOG_FORMAT_SIMPLE" "$LOG_DATE_FORMAT_SIMPLE"; }
# Print empty lines before and on the end of Error+ logs
function ERROR     { echo; LOG_HANDLER_DEFAULT "$@"; echo; exit 1;}
function CRITICAL  { echo; LOG_HANDLER_DEFAULT "$@"; echo; exit 1;}
function ALERT     { echo; LOG_HANDLER_DEFAULT "$@"; echo; exit 1;}
function EMERGENCY { echo; LOG_HANDLER_DEFAULT "$@"; echo; exit 1;}

#######################################
# All log levels call this handler (by default),
# so this is a great place to put any standard logging behavior
# Globals:
#   LOG_FORMAT
#   LOG_DATE_FORMAT
# Arguments:
#   $@ - Message for log. Default ""
#   log_format - Log string format. Default "$LOG_FORMAT"
#   log_date_format - Date and time string format. Default "$LOG_DATE_FORMAT"
#   additional_logging_func_trace_lvl - Remove from function trace last function name. Default - 0
#######################################
function LOG_HANDLER_DEFAULT {
    # $1 - message
    local log_format=${2:-"$LOG_FORMAT"}
    local log_date_format=${3:-"$LOG_DATE_FORMAT"}
    local additional_logging_func_trace_lvl=${4:-"0"}
    # As result get "function_with_trouble <- some_func <- main"
    local logging_func_trace_lvl=2 # LOG_HANDLER_DEFAULT + logging_function
    logging_func_trace_lvl=$((logging_func_trace_lvl + additional_logging_func_trace_lvl))
    local FUNC_TRACE="${FUNCNAME[*]:logging_func_trace_lvl}" # Get functions in trace. Skip first two functions that used for logging
    local func_trace=${FUNC_TRACE// / <- } # Replace ' ' to ' <- '.

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
    formatted_log="$(FORMAT_LOG "$lvl" "$func_trace" "$log_format" "$log_date_format" "$@")"

    LOG_HANDLER_COLORTERM "$lvl" "$formatted_log"
    LOG_HANDLER_LOGFILE "$lvl" "$formatted_log"
}

#######################################
# Outputs a log formatted by provided log and date formats.
# Globals:
#   LOG_FORMAT
#   LOG_DATE_FORMAT
# Arguments:
#   level - Log level. Default "DEBUG"
#   func_trace - Function trace. Default ""
#   formatted_log - Log string format. Default "$LOG_FORMAT"
#   date_format - Date and time string format. Default "$LOG_DATE_FORMAT"
#   log - Message for log. Default ""
#######################################
function FORMAT_LOG {
    local level="${1:-"DEBUG"}"
    local func_trace="${2:-""}"
    local formatted_log="${3:-"$LOG_FORMAT"}"
    local date_format="${4:-"$LOG_DATE_FORMAT"}"
    local log="$5"

    local pid=$$
    local date
    date="$(date "$date_format")"

    formatted_log="${formatted_log/'%MESSAGE'/$log}"
    formatted_log="${formatted_log/'%LEVEL'/$level}"
    formatted_log="${formatted_log/'%PID'/$pid}"
    formatted_log="${formatted_log/'%DATE'/$date}"
    formatted_log="${formatted_log/'%FUNC_TRACE'/$func_trace}"
    echo "$formatted_log"
}

#######################################
# Outputs a log to the stdout, colorized using the LOG_COLOR configurables
# Globals:
#   LOG_COLOR_DEBUG
#   LOG_COLOR_INFO
#   LOG_COLOR_NOTICE
#   LOG_COLOR_WARNING
#   LOG_COLOR_ERROR
#   LOG_COLOR_CRITICAL
#   LOG_COLOR_ALERT
#   LOG_COLOR_EMERGENCY
#   RESET_COLOR
# Arguments:
#   level - Log level.
#   log - Message for log.
#######################################
function LOG_HANDLER_COLORTERM {
    local level="$1"
    local log="$2"
    local color_variable="LOG_COLOR_$level"
    local color="${!color_variable}"
    log="$color$log$RESET_COLOR"
    echo -e "$log"
}

#######################################
# Appends a log to the configured logfile
# Globals:
#   None
# Arguments:
#   level - Log level.
#   log - Message for log.
#######################################
function LOG_HANDLER_LOGFILE {
    local level="$1"
    local log="$2"
    local log_path

    log_path="$(dirname "$LOGFILE")"
    [ -d "$log_path" ] || mkdir -p "$log_path"
    echo "$log" >> "$LOGFILE"
}

#######################################
# Run command into wrapper, that print command output only when:
# - error happened
# - VERBOSE_LVL=DEBUG
# Globals:
#   LOG_LVL
# Arguments:
#   command - command that should be executed inside wrapper
#   bash_opts - additional shell options
#   fail_on_err - interpret not 0 exit code as error or not.
#                 Boolean. By default - true.
#   enable_log_timeout - Print all logs for command after timeout.
#                        Useful for non-DEBUG log levels.
#                        By default - 300 seconds (5 min)
# Outputs:
#   Writes progress status
#######################################
function run_cmd {
    local command="$1"
    local bash_opts="${2-""}"
    local fail_on_err="${3-true}"
    local enable_log_timeout="${4-300}" # By default - 300 seconds (5 min)

    local bash="/usr/bin/env bash ${bash_opts}"

    # STDERR prints by default
    # Log STDOUT and continue
    if [ "$LOG_LVL" == "DEBUG" ]; then
        DEBUG "Output from '$command'" "" "" 1
        ${bash} -x -c "$command"
        exit_code=$?

        [ ${exit_code} != 0 ] && [ "$fail_on_err" = true ] && ERROR "Execution of '$command' failed" "" "" 1
        return
    fi

    # shellcheck disable=SC2091
    $(${bash} -x -c "$command" >/tmp/log) &
    proc_pid=$!

    # Print logs in non-debug
    seconds=0
    until [ ! -d /proc/$proc_pid ]; do
        if [[ $seconds -gt $enable_log_timeout ]]; then
            WARNING "Command '$command' executing more than ${seconds}s. \
Start print out exist and future output for this command." "%DATE - %MESSAGE" "+%T" 1
            tail -f -n +0 /tmp/log &
            tail_pid=$!
            break
        fi
        sleep 1
        ((seconds++))
    done

    wait $proc_pid
    exit_code=$?
    # `> /dev/null 2>&1 &` and `wait` need to suppress "Terminated' message
    kill $tail_pid > /dev/null 2>&1 &
    wait

    [ ${exit_code} != 0 ] && [ "$fail_on_err" = true ] && ERROR "Execution of '$command' failed" "" "" 1
}
