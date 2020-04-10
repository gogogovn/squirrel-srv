#!/bin/sh

for t in $(find ./deployments/kubernetes -type f -name "*.yaml"); do \
    envsubst < $t; \
    echo '---'; \
done > k8s-spec.yaml
