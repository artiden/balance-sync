FROM php:8.5-fpm-alpine

# Install system deps
RUN apk add --no-cache \
    bash \
    git \
    curl \
    zip \
    unzip \
    icu-dev \
    oniguruma-dev \
    libxml2-dev \
    libzip-dev \
    linux-headers

# PHP extensions
RUN docker-php-ext-install pdo pdo_mysql intl mbstring zip xml sockets

# Install Composer
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer

WORKDIR /var/www/html
