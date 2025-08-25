# docker/php/Dockerfile
FROM php:8.2-fpm-alpine

RUN apk add --no-cache --virtual .build-deps \
        $PHPIZE_DEPS \
        linux-headers \
        libzip-dev \
        sqlite-dev \
        libpng-dev \
        libjpeg-turbo-dev \
        freetype-dev \
        postgresql-dev \
        icu-dev \
        libxml2-dev \
        libxslt-dev \
        imagemagick-dev \
        libmemcached-dev \
        rabbitmq-c-dev \
    && apk add --no-cache \
        git \
        libzip \
        sqlite-libs \
        libpng \
        libjpeg-turbo-dev \
        freetype-dev \
        postgresql-libs \
        icu-libs \
        libxml2 \
        libxslt \
        imagemagick \
        libmemcached-libs \
        rabbitmq-c \
    && \
    # Install PECL extensions
    pecl install redis memcached imagick amqp && \
    docker-php-ext-enable redis memcached imagick amqp && \
    # Configure GD
    docker-php-ext-configure gd --with-jpeg --with-freetype && \
    # Install core extensions
    docker-php-ext-install -j$(nproc) \
        bcmath exif gd intl opcache pcntl pdo_mysql pdo_pgsql pdo_sqlite soap sockets xsl zip \
    && apk del .build-deps

# Set a non-root user
RUN addgroup -g 1000 laravel && adduser -u 1000 -G laravel -s /bin/sh -D laravel
USER laravel
WORKDIR /var/www/html

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/local/bin/composer
