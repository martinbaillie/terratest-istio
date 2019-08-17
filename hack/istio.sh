#!/usr/bin/env sh
set -e

# Needs:    golang 1.11+, jq, kubectl, helm, curl
# Setup:    istio.sh setup
# Teardown: istio.sh teardown

# Default to the version of the repo Go module.
: "${ISTIO_VERSION:="gomod"}"

prepare() {
    if [ "${ISTIO_VERSION}" = "gomod" ]; then
        # Change directory to Istio module and use it
        cd "$(go mod download -json | \
            jq -r 'select([.Path == "istio.io/istio"] | any) | .Dir')"
    else
        # Use the specified Istio version
        cd "$(mktemp -d /tmp/istio.XXXXXX)"
        if ! (git clone --depth 1 --branch "${ISTIO_VERSION}" \
            https://github.com/istio/istio) >&2
        then
            >&2 echo "Istio download failed"
            exit 1
        fi
        cd istio
        echo "Testing against specific Istio version: ${ISTIO_VERSION}"
    fi

    # Ensure cluster is ready
    if ! kubectl cluster-info > /dev/null; then
        >&2 echo "Kubectl failed to talk to cluster"
        >&2 echo "Check kube config then run this script again"
        exit 1
    fi
}

setup() {
    prepare

    # Create namespace
    if ! kubectl get namespace istio-system > /dev/null 2>&1; then
        kubectl create namespace istio-system
    fi

    # Install Istio CRDs
    if echo "${ISTIO_VERSION}" | grep -qe '1\.0\..*'; then
        # 1.0 stream
        kubectl apply -f \
            install/kubernetes/helm/istio/templates/crds.yaml
        sleep 10
    elif echo "${ISTIO_VERSION}" | grep -qe '1\.1\..*'; then
        # 1.1 stream
        helm template install/kubernetes/helm/istio-init \
            --name istio-init \
            --namespace istio-system | \
            kubectl apply -f -
        kubectl wait job/istio-init-crd-10 \
            -n istio-system \
            --for=condition=complete \
            --timeout 5m
        kubectl wait job/istio-init-crd-11 \
            -n istio-system \
            --for=condition=complete \
            --timeout 5m
    elif echo "${ISTIO_VERSION}" | grep -qe '1\.2\..*'; then
        # 1.2 stream
        helm template install/kubernetes/helm/istio-init \
            --name istio-init \
            --namespace istio-system | \
            kubectl apply -f -
        kubectl wait job/istio-init-crd-10 \
            -n istio-system \
            --for=condition=complete \
            --timeout 5m
        kubectl wait job/istio-init-crd-11 \
            -n istio-system \
            --for=condition=complete \
            --timeout 5m
        kubectl wait job/istio-init-crd-12 \
            -n istio-system \
            --for=condition=complete \
            --timeout 5m
    else
        # 1.3+ streams (untagged release/master)
        helm template install/kubernetes/helm/istio-init \
            --name istio-init \
            --namespace istio-system | \
            kubectl apply -f -
        kubectl wait job/istio-init-crd-10-master-latest-daily \
            -n istio-system \
            --for=condition=complete \
            --timeout 5m
        kubectl wait job/istio-init-crd-11-master-latest-daily \
            -n istio-system \
            --for=condition=complete \
            --timeout 5m
        kubectl wait job/istio-init-crd-12-master-latest-daily \
            -n istio-system \
            --for=condition=complete \
            --timeout 5m
    fi

    # Install Istio
    helm template install/kubernetes/helm/istio \
        --name istio \
        --namespace istio-system \
        --set "mixer.telemetry.enabled=false" \
        --set "prometheus.enabled=false" \
        --set "grafana.enabled=false" \
        --set "tracing.enabled=false" \
        --set "kiali.enabled=false" \
        --set "gateways.istio-ingressgateway.enabled=false" \
        --set "global.proxy.resources.requests.cpu=50m" \
        --set "pilot.resources.requests.cpu=250m" \
        --values install/kubernetes/helm/istio/values-istio-demo-auth.yaml | \
        kubectl apply -f -
    kubectl wait deployment/istio-pilot \
        -n istio-system \
        --for=condition=available \
        --timeout 5m

    # Install BookInfo sample
    kubectl label namespace default istio-injection=enabled --overwrite
    kubectl apply -f samples/bookinfo/platform/kube/bookinfo.yaml
    kubectl wait deployment/details-v1 --for=condition=available --timeout 5m
    kubectl wait deployment/reviews-v1 --for=condition=available --timeout 5m
    kubectl wait deployment/reviews-v2 --for=condition=available --timeout 5m
    kubectl wait deployment/reviews-v3 --for=condition=available --timeout 5m
    kubectl wait deployment/productpage-v1 \
        --for=condition=available \
        --timeout 5m

    # TODO: figure out why this check doesn't work on the Istio 1.0 or 1.1 stream
    if ! (echo "${ISTIO_VERSION}" | grep -qe '1\.[01]\..*'); then
        kubectl exec -it "$(kubectl get pod -l app=ratings \
            -o jsonpath='{.items[0].metadata.name}')" \
            -c ratings -- \
            curl productpage:9080/productpage | grep -qo "<title>.*</title>"
    fi
}

teardown() {
    prepare

    # Delete BookInfo sample
    kubectl label namespace default istio-injection-
    kubectl delete -f samples/bookinfo/platform/kube/bookinfo.yaml

    # Delete Istio
    helm template install/kubernetes/helm/istio \
        --name istio \
        --namespace istio-system \
        --set "mixer.telemetry.enabled=false" \
        --set "prometheus.enabled=false" \
        --set "grafana.enabled=false" \
        --set "tracing.enabled=false" \
        --set "kiali.enabled=false" \
        --set "gateways.istio-ingressgateway.enabled=false" \
        --set "global.proxy.resources.requests.cpu=50m" \
        --set "pilot.resources.requests.cpu=250m" \
        --values install/kubernetes/helm/istio/values-istio-demo-auth.yaml | \
        kubectl delete -f -

    # Delete CRDs
    #kubectl delete -f install/kubernetes/helm/istio-init/files

    # Delete NS
    kubectl delete namespace istio-system
}

case "$1" in
    "setup") setup ;;
    "teardown") teardown ;;
    *) >&2 echo "Usage: $0 [setup|teardown]" && exit 1 ;;
esac
