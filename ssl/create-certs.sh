#! /bin/bash

set -euo pipefail

NAME=$1
WEBHOOK_NS=$2
WEBHOOK_SVC="${NAME}-webhook"

root="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null 2>&1 && pwd )"
ssl=$root/ssl
cd $root

# Create certs for our webhook
openssl genrsa -out $ssl/webhookCA.key 2048
openssl req -new -key $ssl/webhookCA.key -subj "/CN=${WEBHOOK_SVC}.${WEBHOOK_NS}.svc" -out $ssl/webhookCA.csr
openssl x509 -req -days 3650 -in $ssl/webhookCA.csr -signkey $ssl/webhookCA.key -out $ssl/webhook.crt

# Create certs secrets for k8s
kubectl create secret generic \
    ${WEBHOOK_SVC}-certs \
    --from-file=key.pem=$ssl/webhookCA.key \
    --from-file=cert.pem=$ssl/webhook.crt \
    --dry-run -o yaml > $root/deploy/webhook-certs.yaml

# Set the CABundle on the webhook registration
webhook_registration_file=$root/deploy/webhook-registration.yaml
sed -i '' "s/caBundle:.*/caBundle: $(cat $ssl/webhook.crt | base64)/" $webhook_registration_file

# Clean
rm $ssl/webhookCA* && rm $ssl/webhook.crt