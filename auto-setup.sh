#!/bin/sh

# This script runs inside the 'setup' container.
# It reads sites.txt and generates Nginx configs and SSL certificates.

set -e # Exit immediately if a command exits with a non-zero status.

SITES_FILE="/app/sites.txt"
NGINX_SITES_DIR="/etc/nginx/conf.d"
NGINX_SSL_DIR="/etc/nginx/ssl"
MKCERT_CA_DIR="/root/.local/share/mkcert"

echo "[INFO] Starting site setup..."

# --- Trust the local CA ---
# This step is crucial for mkcert to generate trusted certificates.
# We assume the rootCA.pem is mounted from the host.
if [ -f "$MKCERT_CA_DIR/rootCA.pem" ]; then
    echo "[INFO] Found existing Certificate Authority."
else
    echo "[INFO] No local Certificate Authority found. Creating a new one..."
    # This will create the rootCA.pem in the mounted volume
    mkcert -install
    echo "[SUCCESS] New Certificate Authority created in ./mkcert-ca"
    echo "[ACTION REQUIRED] You must now trust the new CA on your host machine."
    echo "Run this on your host: 'sudo cp ./mkcert-ca/rootCA.pem /usr/local/share/ca-certificates/mkcert_rootCA.crt && sudo update-ca-certificates'"
    echo "And import it into Firefox if you use it."
fi


# Check for input file
if [ ! -f "$SITES_FILE" ]; then
    echo "[ERROR] Input file '$SITES_FILE' not found. Exiting."
    exit 1
fi

echo "[INFO] Reading domains from $SITES_FILE..."

# Process each site from the input file
while IFS= read -r site_url || [ -n "$site_url" ]; do
    if [ -z "$site_url" ]; then
        continue
    fi

    protocol=$(echo "$site_url" | grep -o '^[a-z]*')
    # Extract domain, removing protocol and (php-xx) part
    domain=$(echo "$site_url" | sed -e 's|^[^/]*//||' -e 's/(.*//')
    project_name=$(echo "$domain" | cut -d'.' -f1)
    conf_file="$NGINX_SITES_DIR/$domain.conf"

    # Extract PHP version, with a default value
    php_version_raw=$(echo "$site_url" | grep -o '(.*)' | tr -d '()')
    if [ -n "$php_version_raw" ]; then
        PHP_SERVICE=$php_version_raw
    else
        PHP_SERVICE="php-82" # Default PHP version
    fi

    echo "[INFO] Processing domain: $domain"

    if [ "$protocol" = "https" ]; then
        # --- HTTPS Setup ---
        echo "[INFO] Generating SSL certificate for $domain..."
        mkcert -cert-file "$NGINX_SSL_DIR/$domain.crt" -key-file "$NGINX_SSL_DIR/$domain.key" "$domain"

        echo "[INFO] Generating HTTPS Nginx config for $domain..."
        cat > "$conf_file" <<EOF
# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name $domain;
    return 301 https://\$host\$request_uri;
}

# Serve content over HTTPS
server {
    listen 443 ssl;
    server_name $domain;
    root /var/www/html/$project_name/public;
    index index.php;

    ssl_certificate /etc/nginx/ssl/$domain.crt;
    ssl_certificate_key /etc/nginx/ssl/$domain.key;

    location / { try_files \$uri \$uri/ /index.php?\$query_string; }
    location ~ \.php$ {
        fastcgi_pass $PHP_SERVICE:9000;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;
    }
}
EOF
    else
        # --- HTTP Setup ---
        echo "[INFO] Generating HTTP Nginx config for $domain..."
        cat > "$conf_file" <<EOF
server {
    listen 80;
    server_name $domain;
    root /var/www/html/$project_name/public;
    index index.php;

    location / { try_files \$uri \$uri/ /index.php?\$query_string; }
    location ~ \.php$ {
        fastcgi_pass $PHP_SERVICE:9000;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME \$document_root\$fastcgi_script_name;
    }
}
EOF
    fi
    echo "[SUCCESS] Setup complete for $domain."
done < "$SITES_FILE"

echo "[SUCCESS] All sites have been configured. The setup container will now exit."
