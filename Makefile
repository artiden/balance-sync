DOCKER_COMPOSE = docker compose

COMPOSER = $(DOCKER_COMPOSE) exec laravel-app composer
PHP = $(DOCKER_COMPOSE) exec laravel-app php
ARTISAN = $(PHP) artisan
SCHEDULER_ARTISAN = $(DOCKER_COMPOSE) exec laravel-scheduler php artisan

init:
	$(MAKE) build
	$(MAKE) up
	$(MAKE) composer-install
	$(MAKE) migrate
	$(MAKE) seed
	$(MAKE) scheduler-start
	@echo "Initialized."

up:
	$(DOCKER_COMPOSE) up -d

build:
	$(DOCKER_COMPOSE) build

down:
	$(DOCKER_COMPOSE) down

restart:
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) up -d

scheduler-start:
	$(SCHEDULER_ARTISAN) schedule:work

composer-install:
	$(COMPOSER) install

migrate:
	$(ARTISAN) migrate

migrate-fresh:
	$(ARTISAN) migrate:fresh

seed:
	$(ARTISAN) db:seed

tinker:
	$(ARTISAN) tinker

logs:
	$(DOCKER_COMPOSE) logs -f

bash:
	$(DOCKER_COMPOSE) exec laravel-app bash

cache-clear:
	$(ARTISAN) optimize:clear
