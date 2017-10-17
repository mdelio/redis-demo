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

run "" # Wait for input

desc "Create a couple of namespaces for the demo"
run "cat $(relative namespaces.yaml); kubectl apply -f $(relative namespaces.yaml)"


desc "Setup a External Load-Balancer Service for the Frontend (in demo-prod)"
run "cat $(relative frontend-prod-svc.yaml); kubectl apply -f $(relative frontend-prod-svc.yaml)"

desc "...and another for the EVIL Frontend (in demo-dev)"
run "kubectl apply -f $(relative frontend-dev-svc.yaml)"


desc "Deploy a redis instance to demo-prod"
run "cat $(relative redis.yaml); kubectl apply -f $(relative redis.yaml)"

desc "Setup a cluster service for the redis instance"
run "cat $(relative redis-svc.yaml); kubectl apply -f $(relative redis-svc.yaml)"

desc "Setup our frontend pod in the demo-prod namespace"
run "cat $(relative frontend-prod.yaml); kubectl apply -f $(relative frontend-prod.yaml)"

PROD_IP=$(kubectl --namespace=demo-prod describe svc/frontend-svc | grep "LoadBalancer Ingress:" | cut -f2)
desc "Now let's open up our browser and see our app: http://${PROD_IP}"
run "" # Wait for Input

desc "Setup an EVIL frontend pod in the demo-dev namespace"
run "cat $(relative frontend-dev.yaml); kubectl apply -f $(relative frontend-dev.yaml)"

DEV_IP=$(kubectl --namespace=demo-dev describe svc/frontend-svc | grep "LoadBalancer Ingress:" | cut -f2)
desc "Now let's open up our browser and go to the EVIL app: http://${DEV_IP}"
run "" # Wait for Input


desc "Create a policy that blocks the EVIL Frontend from reaching redis"
run "cat $(relative redis-policy.yaml); kubectl apply -f $(relative redis-policy.yaml)"

desc "Now let's check on our two services again"
run "" # Wait for Input