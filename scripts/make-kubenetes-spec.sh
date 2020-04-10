#!/bin/sh

for t in $(find ./deploy/k8s -type f -name "*.yaml"); do \
    envsubst < $t; \
    echo '---'; \
done > k8s-spec.yaml
