#!/usr/bin/env bash
set -eu

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
DESIRED_ISTIO_VERSION=${DESIRED_ISTIO_VERSION:-"1.6.0"}


istioctl_version="$(istioctl version --remote=false)"
if [[ ${istioctl_version} != "${DESIRED_ISTIO_VERSION}" ]]; then
  echo "Please install version ${DESIRED_ISTIO_VERSION} of istioctl: https://github.com/istio/istio/releases/tag/${DESIRED_ISTIO_VERSION}" >&2
  exit 1
fi

echo "generating Istio resource definitions to ${SCRIPT_DIR}/tmp/generated-istio.yaml ..." >&2
mkdir -p "${SCRIPT_DIR}/../../istio"
mkdir -p "${SCRIPT_DIR}/tmp"
istioctl manifest generate -f "${SCRIPT_DIR}/istio-values-gen.yaml" "$@" > "${SCRIPT_DIR}/tmp/generated-istio.yaml"

ytt --ignore-unknown-comments \
  -f "${SCRIPT_DIR}/ytt-data-values.yaml" \
  -f "${SCRIPT_DIR}/tmp/generated-istio.yaml" \
  -f "${SCRIPT_DIR}/overlays"
