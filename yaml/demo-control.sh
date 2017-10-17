#!/bin/bash
# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

. $(dirname ${BASH_SOURCE})/util.sh

desc "Create a couple of namespaces for the demo"
run "cat $(relative namespaces.yaml)"
run "kubectl apply -f $(relative namespaces.yaml)"

desc "Setup a External Load-Balancer Service for the Frontend"
run "cat $(relative frontend-prod-svc.yaml)"
run "kubectl apply -f $(relative frontend-prod-svc.yaml)"

desc "Setup a External Load-Balancer Service for the EVIL Frontend"
run "cat $(relative frontend-dev-svc.yaml)"
run "kubectl apply -f $(relative frontend-dev-svc.yaml)"


desc "Deploy a redis instance in the demo-prod namespace"
run "cat $(relative redis.yaml)"
run "kubectl apply -f $(relative redis.yaml)"

desc "Setup an internal Kubernetes service for the redis instance"
run "cat $(relative redis-svc.yaml)"
run "kubectl apply -f $(relative redis-svc.yaml)"


desc "Setup a frontend pod in the demo-prod namespace"
run "cat $(relative frontend-prod.yaml)"
run "kubectl apply -f $(relative frontend-prod.yaml)"


desc "Setup an EVIL frontend pod in the demo-dev namespace"
run "cat $(relative frontend-dev.yaml)"
run "kubectl apply -f $(relative frontend-dev.yaml)"


desc "Apply a Network Policy that blocks the EVIL Frontend from reaching Redis"
run "cat $(relative redis-policy.yaml)"
run "kubectl apply -f $(relative redis-policy.yaml)"