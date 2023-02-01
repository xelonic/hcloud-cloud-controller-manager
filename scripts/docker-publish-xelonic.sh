#!/bin/bash

set -euo pipefail
set -x

print_usage() {
    echo
    echo "$(basename $0) <version>"
    echo
    echo "version: version to publish"
}

if (( $# < 1 ));
then
    echo "missing parameter."
    print_usage
    exit 1
fi

version=$1; shift

local_image_name="hcloud-cloud-controller-manager:${version}-root-server-support"
remote_image_name="docker.io/xelonic/${local_image_name}"


docker build -t "${local_image_name}" .
docker tag "${local_image_name}" "${remote_image_name}"
docker push "${remote_image_name}"
