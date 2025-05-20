#!/bin/sh
# Script to generate runtime config.js file in a writable location

# Create config directory if it doesn't exist
mkdir -p /tmp/config

# Generate a new config.js file with environment variables
cat > /tmp/config/config.js << EOF
// Runtime configuration - these values are replaced at deployment time
window.RUNTIME_CONFIG = {
  AUTH_URL: "${VITE_AUTH_URL:-http://localhost:5556/dex}",
  API_URL: "${VITE_API_URL:-http://localhost:8880}",
  REDIRECT_URI: "${VITE_REDIRECT_URI:-http://localhost:5173}"
};
EOF

# Execute nginx
exec "$@"
