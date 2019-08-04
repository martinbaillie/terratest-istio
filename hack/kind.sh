#!/usr/bin/env sh
set -e

# Needs:    golang 1.11+
# Setup:    eval $(kind.sh setup)
# Teardown: eval $(kind.sh teardown)

teardown() {
    if kind --help >/dev/null && \
        (kind get clusters | grep -q terratest-istio)
    then
        kind delete cluster --name=terratest-istio >/dev/null
        echo "unset KUBECONFIG"
    fi
}

setup() {
    # Install KinD if missing
    if ! kind --help > /dev/null 2>&1
    then
        #if ! (>&2 go get -u sigs.k8s.io/kind)
        #https://github.com/kubernetes-sigs/kind/issues/508
        if ! (cd /tmp && git clone https://github.com/kubernetes-sigs/kind && \
                cd kind && git checkout v0.2.1 && go install) >&2
        then
            >&2 echo "KinD installation failed"
            >&2 echo "Install it manually then run this script again"
            exit 1
        fi
    # Otherwise teardown existing
    else
        teardown
    fi

    # Create a new cluster
    if ! (kind create cluster --name=terratest-istio) > /dev/null
    then
        >&2 echo "Could not setup KinD environment"
        >&2 echo "Check KinD install and try again"
        exit 1
    fi

    echo "export KUBECONFIG=$(kind get kubeconfig-path --name=terratest-istio)"
}

case "$1" in
    "setup") setup ;;
    "teardown") teardown ;;
    *) >&2 echo "Usage: $0 [setup|teardown]" && exit 1 ;;
esac
