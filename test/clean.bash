#!/bin/bash

if [[ -z "$1" ]]; then
    echo "Please specify a namespace to clean as an argument to this script"
    exit 1
fi

NAMESPACE="$1"
echo "Cleaning namespace $NAMESPACE"

iofogctl delete all -n "$NAMESPACE"
kubectl delete all --all -n "$NAMESPACE"
kubectl delete clusterrolebindings kubelet
kubectl delete crd iofogs.k8s.iofog.org
kubectl delete ns "$NAMESPACE"
iofogctl delete namespace "$NAMESPACE"