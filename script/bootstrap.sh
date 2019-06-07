#!/usr/bin/env sh
#
# bootstrap.sh will check for and install any dependencies we have for building and using iofogctl
#
# Usage: ./bootstrap.sh
#
#

set -e

# Import our helper functions
. script/utils.sh

prettyTitle "Installing iofogctl Dependencies"
echo

# What platform are we on?
OS=$(uname -s | tr A-Z a-z)
HELM_VERSION=2.13.1
K8S_VERSION=1.13.4

# Check whether Brew is installed
# TODO: Current installation method is macos centric, make it work for linux too.
if ! checkForInstallation "brew"; then
    echoInfo " Attempting to install Brew"
    /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
fi

#
# All our Go related stuff
#

# Is go installed?
if ! checkForInstallation "go"; then
    echoInfo " Attempting to install 'golang'"
    brew install go
fi

# Is dep installed?
if ! checkForInstallation "dep"; then
    echoInfo " Attempting to install 'go dep'"
    go get -u github.com/golang/dep/cmd/dep
fi

# Is go-junit-report installed?
if ! checkForInstallation "go-junit-report"; then
    echoInfo " Attempting to install 'go-junit-report'"
    go get -u jstemmer/go-junit-report
fi


#
# All our Kubernetes related stuff
#

# Is helm installed?
if ! checkForInstallation "helm"; then
    echoInfo " Attempting to install helm"
    brew install kubernetes-helm
fi

# Is kubernetes-cli installed?
if ! checkForInstallation "kubectl"; then
    echoInfo " Attempting to install kubernetes-cli"
    brew install kubernetes-cli
fi