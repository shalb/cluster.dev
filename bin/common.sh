#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

#######################################
# Writes information about used software
# Globals:
#   None
# Arguments:
#   None
# Outputs:
#   Writes software versions
#######################################
function output_software_info {
    DEBUG "Software installed information:"
    DEBUG "helm: $(helmfile -v)"
    DEBUG "kubectl: $(kubectl version --client)"
    DEBUG "git: $(git --version)"
    DEBUG "AWS CLI: $(aws --version) "
    DEBUG "YAMLTOENV: $(yamltoenv)"
}

#######################################
# Detect the cloud provider which actually run the script
# It could be GitHub Action, Gitlab Pipeline, Bitbucket Pipeline
# Globals:
#   None
# Arguments:
#   Environment values which are set by CI/CD platform
# Outputs:
#   GIT_PROVIDER - platform for git hosting
#   GIT_REPO_NAME - repo in "user/repository" format
#######################################
function detect_git_provider {
        # Check if it is GitHub
        # https://help.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables#default-environment-variables
        if [ ! -z ${GITHUB_REPOSITORY+h} ];
                then DEBUG "GITHUB_REPOSITORY variable is set to $GITHUB_REPOSITORY"
                    INFO "Assuming we are running GitHub Action";
                    readonly GIT_PROVIDER="github"
                    readonly GIT_REPO_NAME=$GITHUB_REPOSITORY
                    readonly GIT_REPO_ROOT=$GITHUB_WORKSPACE
            else
                DEBUG "GITHUB_REPOSITORY variable is NOT set assuming we are NOT running GitHub Action";
        fi

        # Check if it is GitLab
        # https://docs.gitlab.com/ee/ci/variables/predefined_variables.html
        if [ ! -z ${CI_PROJECT_PATH+l} ];
                then DEBUG "CI_PROJECT_PATH variable is set to $CI_PROJECT_PATH"
                    INFO "Assuming we are running GitLab Pipeline";
                    readonly GIT_PROVIDER="gitlab"
                    readonly GIT_REPO_NAME=$CI_PROJECT_PATH
                    readonly GIT_REPO_ROOT=$CI_PROJECT_DIR
            else
                DEBUG "CI_PROJECT_PATH variable is NOT set assuming we are NOT running GitLab Pipeline";
        fi

        # Check if it is Bitbucket
        # https://confluence.atlassian.com/bitbucket/variables-in-pipelines-794502608.html#Variablesinpipelines-Defaultvariables
        if [ ! -z ${BITBUCKET_GIT_HTTP_ORIGIN+b} ];
                then DEBUG "BITBUCKET_GIT_HTTP_ORIGIN variable is set to $BITBUCKET_GIT_HTTP_ORIGIN"
                    INFO "Assuming we are running Bitbucket Pipeline";
                    readonly GIT_PROVIDER="bitbucket"
                    readonly GIT_REPO_NAME=$(echo $BITBUCKET_GIT_HTTP_ORIGIN | sed -e 's/http:\/\/bitbucket.org\///g')
                    readonly GIT_REPO_ROOT=$BITBUCKET_CLONE_DIR
            else
                DEBUG "BITBUCKET_GIT_HTTP_ORIGIN variable is NOT set assuming we are NOT running Bitbucket Pipeline";
        fi

        # Output final results with required variables set
        INFO "GIT_PROVIDER is set for: $GIT_PROVIDER"
        INFO "GIT_REPO_NAME is set for: $GIT_REPO_NAME"
        INFO "GIT_REPO_ROOT is set for: $GIT_REPO_ROOT"
}


#######################################
# Generate a unique name for particular cluster domains, state bucketes, etc..
# Globals:
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name - from yaml file
#   GIT_REPO_NAME - from provider detection
# Outputs:
#   CLUSTER_FULLNAME - naming for state buckets and other unique names
#######################################
function set_cluster_fullname {
    local cluster_name=$1
    local git_repo_name=$2
    local CLUSTER_FULLNAME=""

    # Define CLUSTER_FULLNAME which will be used in state files
    CLUSTER_FULLNAME=$cluster_name-$(echo "$git_repo_name" | awk -F "/" '{print$1}')
    # make sure it is not larger than 63 symbols and lowercase
    CLUSTER_FULLNAME=$(echo "$CLUSTER_FULLNAME" | cut -c 1-63 | awk '{print tolower($0)}')

    INFO "CLUSTER_FULLNAME is set for: $CLUSTER_FULLNAME"
    # shellcheck disable=SC2034
    FUNC_RESULT="${CLUSTER_FULLNAME}"
}

#######################################
# Convert string value to terraform list
# Globals:
# Arguments:
#   string in format: "some1, some2, someT"
# Outputs:
#   list in ["some1", "some2", "someT"]
#######################################
function to_tf_list {
    local source_string=$1
    local result=""
        if  [[ ! -z $source_string ]] ; then
        IFS=', ' read -r -a array <<< "$source_string";
        for element in "${array[@]}"
            do
                result+=\ "\"$element\"",
            done
                result="[${result::-1} ]"
            else
            result="[]"
        fi
        echo "$result"
}
