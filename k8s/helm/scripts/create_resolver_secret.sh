#!/bin/bash

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Get the repository root (3 levels up from the script location)
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
RESOLVER_FILE="$REPO_ROOT/resolvers_settings.yaml"

# Get secrets.sh location (either from argument or use same directory as script)
SECRETS_DIR="${1:-$SCRIPT_DIR}"
SECRETS_FILE="$SECRETS_DIR/secrets.sh"

# Encode resolvers_settings.yaml to base64 (compatible with both Linux and macOS)
if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS uses different base64 options
  ENCODED_RESOLVER=$(cat "$RESOLVER_FILE" | base64)
else
  # Linux version
  ENCODED_RESOLVER=$(cat "$RESOLVER_FILE" | base64 -w 0)
fi

# Check if secrets.sh exists
if [ -f "$SECRETS_FILE" ]; then
  # Check if ISSUER_RESOLVER_FILE already exists in the file
  if grep -q "ISSUER_RESOLVER_FILE=" "$SECRETS_FILE"; then
    # Replace the existing entry - compatible with both Linux and macOS
    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' "s|ISSUER_RESOLVER_FILE=.*|ISSUER_RESOLVER_FILE=\"$ENCODED_RESOLVER\"|" "$SECRETS_FILE"
    else
      sed -i "s|ISSUER_RESOLVER_FILE=.*|ISSUER_RESOLVER_FILE=\"$ENCODED_RESOLVER\"|" "$SECRETS_FILE"
    fi
  else
    # Add the new entry at the end of the file
    echo "ISSUER_RESOLVER_FILE=\"$ENCODED_RESOLVER\"" >> "$SECRETS_FILE"
  fi
else
  # Create a new secrets.sh file
  echo "#!/bin/bash" > "$SECRETS_FILE"
  echo "ISSUER_RESOLVER_FILE=\"$ENCODED_RESOLVER\"" >> "$SECRETS_FILE"
  chmod +x "$SECRETS_FILE"
fi

echo "Base64 encoded resolvers_settings.yaml has been added to $SECRETS_FILE as ISSUER_RESOLVER_FILE"
