# Balance Update System — Laravel + Golang + RabbitMQ

This project is a two-part system demonstrating periodic balance updates, event publishing, message consumption, and data synchronization between services.
It consists of:

* **Laravel application** — manages users and their balances, periodically modifies them, and publishes updates to RabbitMQ.
* **Golang microservice** — subscribes to RabbitMQ events, stores received balance changes in its own database, and maintains an internal cache.

Docker and `make` are used for full automation.

---

## 🧩 System Overview

### **Laravel (Part 1)**

Laravel is responsible for:

* Creating a **users** table (~1000 records) and a **balances** table.
* Providing **seeders** to generate test users and initial balances.
* Running a **scheduled job every 10 seconds** (configurable) that:

  * selects a random group of users,
  * modifies their balances,
  * saves results to the database,
  * publishes updates to **RabbitMQ**.

### **Golang Service (Part 2)**

A standalone microservice that:

* Connects to **RabbitMQ** and receives all balance-update events.
* Stores updates in its **own database** (`go_service`).
* Maintains an **in-memory cache**, periodically syncing it to the DB.

> The Go service **has no external HTTP API** — this is intentional, as the assignment does not require any public interface.

### **Cache Logic Note**

The cache implementation is minimal — only what's required by the assignment.
For real-world usage, the logic must support lookups, e.g.:

1. If key exists in cache → return value.
2. If key is missing → try DB:

   * if found → store in cache and return;
   * if not found → return `nil` or throw an error depending on requirements.

---

## 🚀 How the System Works (Short Summary)

1. **Laravel Scheduler** randomly updates users’ balances every 10 seconds.
2. Each update is **written to the Laravel DB**.
3. Laravel publishes messages describing the updates to **RabbitMQ**.
4. The **Go service** receives these messages.
5. The Go service stores them in the **Go MySQL database** and updates its **in-memory cache**.
6. Periodic cache synchronization ensures consistency even under:

   * high event rate,
   * out-of-order messages,
   * duplicate messages.

The system remains stable despite heavy updates and message bursts.

---

## 🐳 Requirements

* **Docker**
* **Docker Compose**
* **make**

---

## 📦 Project Startup

### **First run**

```bash
make init
```

This will automatically:

1. Build all containers
2. Start Docker Compose
3. Install Laravel dependencies
4. Run migrations
5. Seed initial data (1000+ users)
6. Start Laravel scheduler
7. Launch all services (Laravel, Go, RabbitMQ, MySQL)

After this, balance updates will begin appearing in the Go service database (`balances` table).

---

## ▶️ Everyday Commands

Start containers:

```bash
make up
```

Rebuild:

```bash
make build
```

Shutdown:

```bash
make down
```

Restart:

```bash
make restart
```

Run Laravel scheduler manually:

```bash
make scheduler-start
```

---

## ⏱ Changing the Update Interval

Balance update frequency is defined in:

```
src/php/routes/console.php
```

Default: **every 10 seconds**
You may modify it as needed.

---

## ⚠️ Notes for Windows Users

Docker Desktop on Windows may have issues with:

* volume mounting,
* file permissions,
* MySQL startup.

If something fails:

1. **Delete Docker volumes**
2. Run the project again

Additionally, MySQL health checks on Windows sometimes require increasing retry counts or timeout values.
The compose file already includes health checks, but you may adjust them if needed.

---

## 🗄 Docker Compose Structure

Includes the following services:

* `laravel-app`
* `laravel-scheduler`
* `laravel-mysql`
* `go-service`
* `go-mysql`
* `rabbitmq`

(See `docker-compose.yml` for details.)

---

## ✔️ Final Notes

* The system is fully automated — just run `make init`.
* RabbitMQ UI is available at **[http://localhost:15672](http://localhost:15672)** (guest/guest).
* Go service contains **no external API**, only internal RabbitMQ and MySQL operations.
* Cache logic is minimal by design and must be extended for real-life usage.

---
