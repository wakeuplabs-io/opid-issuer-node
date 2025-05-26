source ./scripts/secrets.sh
./scripts/create_resolver_secret.sh

helm install "$APP_INSTANCE_NAME" . \
  --create-namespace --set namespace="$NAMESPACE" \
  --set ingressEnabled="$INGRESS_ENABLED" \
  --set uiUsername="$UI_USERNAME" \
  --set uiPassword="$UI_PASSWORD" \
  --set publicIP="$STATIC_IP_NAME" \
  --set uidomain="$UI_DOMAIN" \
  --set apidomain="$API_DOMAIN" \
  --set issuerName="$ISSUERNAME" \
  --set privateKey="$PRIVATE_KEY" \
  --set vaultpwd="$VAULT_PWD" \
  --set issuerResolverFile="$ISSUER_RESOLVER_FILE"
