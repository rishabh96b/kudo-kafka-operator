#!/usr/bin/env bash

set -eux

KUTTL_VERSION=0.6.1
KUBECTL_KUDO_VERSION=${DS_KUDO_VERSION#v}

ARTIFACTS=kuttl-dist

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
KUDO_MACHINE=$(uname -m)
MACHINE=$(uname -m)
if [ "${MACHINE}" == "x86_64" ]; then
  MACHINE=amd64
fi

mkdir -p bin/

# install kuttl, kudo

if [ ! -f "bin/kubectl-kuttl_${KUTTL_VERSION}_${OS}" ]; then
	curl -Lo "bin/kubectl-kuttl_${KUTTL_VERSION}_${OS}" "https://github.com/kudobuilder/kuttl/releases/download/v${KUTTL_VERSION}/kubectl-kuttl_${KUTTL_VERSION}_${OS}_${KUDO_MACHINE}"
	chmod +x "bin/kubectl-kuttl_${KUTTL_VERSION}_${OS}"
fi
ln -sf "./kubectl-kuttl_${KUTTL_VERSION}_${OS}" ./bin/kubectl-kuttl

if [ ! -f "bin/kubectl-kudo_${KUBECTL_KUDO_VERSION}_${OS}" ]; then
	curl -Lo "bin/kubectl-kudo_${KUBECTL_KUDO_VERSION}_${OS}" "https://github.com/kudobuilder/kudo/releases/download/v${KUBECTL_KUDO_VERSION}/kubectl-kudo_${KUBECTL_KUDO_VERSION}_${OS}_${KUDO_MACHINE}"
	chmod +x "bin/kubectl-kudo_${KUBECTL_KUDO_VERSION}_${OS}"
fi
ln -sf "./kubectl-kudo_${KUBECTL_KUDO_VERSION}_${OS}" ./bin/kubectl-kudo

PATH="$(pwd)/bin:${PATH}"

# Test against Kubernetes 1.16
kubectl-kuttl test --config=kuttl-tests.yaml --kind-config=kind/kubernetes-1.16.9.yaml --report=xml --artifacts-dir=${ARTIFACTS}

# Test against Kubernetes 1.17
kubectl-kuttl test --config=kuttl-tests.yaml --kind-config=kind/kubernetes-1.17.5.yaml --report=xml --artifacts-dir=${ARTIFACTS}
