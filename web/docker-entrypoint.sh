#!/bin/sh
# Script to replace environment variables in the runtime config.js file

# Replace values in the config.js file
if [ ! -z "$VITE_AUTH_URL" ]; then
  sed -i "s|AUTH_URL: \"[^\"]*\"|AUTH_URL: \"$VITE_AUTH_URL\"|g" /usr/share/nginx/html/admin/config.js
fi

if [ ! -z "$VITE_API_URL" ]; then
  sed -i "s|API_URL: \"[^\"]*\"|API_URL: \"$VITE_API_URL\"|g" /usr/share/nginx/html/admin/config.js
fi

if [ ! -z "$VITE_REDIRECT_URI" ]; then
  sed -i "s|REDIRECT_URI: \"[^\"]*\"|REDIRECT_URI: \"$VITE_REDIRECT_URI\"|g" /usr/share/nginx/html/admin/config.js
fi

# Execute nginx
exec "$@"
