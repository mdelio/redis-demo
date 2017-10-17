#!/bin/bash

kubectl delete --ignore-not-found=true -f namespaces.yaml

while kubectl get namespace demo-prod >/dev/null 2>&1; do
  kubectl get namespaces
  sleep 1
done

while kubectl get namespace demo-dev >/dev/null 2>&1; do
  kubectl get namespaces
  sleep 1
done

kubectl get namespaces
tmux kill-session -t demo-session >/dev/null 2>&1
