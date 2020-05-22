#!/bin/bash

set -euo pipefail

END=1000
# alltheroutes=$(mktemp)
alltheroutes="/tmp/alles.yml"

for i in $(seq 1 $END); do
  tmp=$(mktemp)
  cat > $tmp <<EOD
---
apiVersion: networking.cloudfoundry.org/v1alpha1
kind: Route
metadata:
 name: cc-route-guid-$i
 annotations: {}
 labels:
   app.kubernetes.io/name: cc-route-guid
   app.kubernetes.io/version: cloud-controller-api-version
   app.kubernetes.io/managed-by: cloudfoundry
   app.kubernetes.io/component: cf-networking
   app.kubernetes.io/part-of: cloudfoundry
   cloudfoundry.org/org_guid: cc-org-guid
   cloudfoundry.org/space_guid: cc-space-guid
   cloudfoundry.org/domain_guid: cc-domain-guid
   cloudfoundry.org/route_guid: cc-route-guid
spec:
  host: hostname-$i
  path: /$i
  url: hostname-$i.apps.example.com/$i
  domain:
    name: apps.example.com
    internal: false
  destinations:
  - weight: 100
    port: 8080
    guid: destination-guid-$i
    selector:
      matchLabels:
        cloudfoundry.org/app_guid: cc-app$i-guid
        cloudfoundry.org/process_type: web
    app:
      guid: cc-app$i-guid
      process:
        type: web
EOD

  cat $tmp >> $alltheroutes
  rm $tmp
done

kubectl apply -f $alltheroutes
