# -*- mode: Python -*-

docker_build('configmatch-dev', '.', dockerfile='Dockerfile')
k8s_yaml(kustomize('./config/default'))
k8s_resource('configma-controller-manager')
