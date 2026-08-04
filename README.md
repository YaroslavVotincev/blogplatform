# Blog Platform

A Go microservice backend for a blogging platform. The system supports user registration and authentication, personal and thematic blogs, posts and rich content, comments, follows, subscriptions, paid access, donations, notifications, file storage, billing, and administrative history logs.

The repository is organized as independent Go modules deployed together with PostgreSQL and RabbitMQ. Clients enter through a single API gateway; services communicate over HTTP for synchronous work and RabbitMQ for asynchronous work.


## Services

| Directory | Responsibility | Main API prefix |
| --- | --- | --- |
| [`api-gateway`](api-gateway) | Public entry point, request proxying, authentication enrichment, and dynamic service discovery | `/api/v1/*` |
| [`auth-service`](auth-service) | Login, JWT creation and validation, roles, account state, and ban state | `/api/v1/auth` |
| [`registration-service`](registration-service) | Signup, confirmation codes, account activation, and signup email events | `/api/v1/registration` |
| [`users-service`](users-service) | Profiles, avatars, wallets, account administration, and user lookup | `/api/v1/users` |
| [`posts-service`](posts-service) | Blogs, posts, content, feeds, follows, subscriptions, payments, donations, goals, and statistics | `/api/v1/blogs`, `/api/v1/posts` |
| [`comments-service`](comments-service) | Create, list, and count comments attached to a parent resource | `/api/v1/comments` |
| [`file-service`](file-service) | Upload, retrieve, and delete files; consumes asynchronous file commands | `/api/v1/files` internally |
| [`notifications-service`](notifications-service) | Consume domain events and expose a user's notification inbox | `/api/v1/notifications` |
| [`email-service`](email-service) | Consume email jobs and deliver transactional email through SMTP.BZ | Worker only |
| [`billing-service`](billing-service) | Robokassa invoices and callbacks, paid-access fulfillment, donations, and TON price lookup | `/api/v1/billing` |
| [`config-service`](config-service) | Store service settings and serve configuration to running services | `/api/v1/config` |
| [`logs-service`](logs-service) | Store and query structured history logs | `/api/v1/logs` |

Each service owns a `cmd/main.go` entry point and follows this layout:

```text
<service>/
├── cmd/                 application entry point
├── internal/            handlers, services, repositories, and workers
├── migrations/          service-owned PostgreSQL migrations, when needed
├── pkg/                 reusable infrastructure within the module
├── Dockerfile
├── go.mod
└── Makefile             convenience targets, when present
```


## Features

### Publishing

- Personal and thematic blogs with draft/public status, URL slugs, categories, avatars, and covers.
- Draft/public posts with tags, rich JSON/HTML content, covers, and content file attachments.
- Feed queries by blog, category, followed blogs, and engagement thresholds.
- Likes, dislikes, anonymous and authenticated views, comments, and aggregate statistics.
- Author goals and blog income history.

### Access and monetization

- Free and paid blog subscriptions.
- Post access modes supporting open, restricted, subscription-based, and paid content.
- Robokassa payment links and payment confirmation callbacks.
- RUB and Toncoin donations.
- Currency-rate lookup for Toncoin pricing.

### Accounts and communication

- Email/password signup with expiring confirmation codes.
- JWT login with optional extended lifetime through.
- User roles, account enable/deletion state, and temporary bans.
- Profiles, avatars, wallet addresses, and RUB wallet balances.
- Asynchronous email, file, notification, and logging workflows.


## Messaging

Usage of RabbitMQ queues:

| Producer | Queue purpose | Consumer |
| --- | --- | --- |
| Registration service | Signup confirmation email | Email service |
| Auth, registration, and posts services | Authentication and domain notifications | Notifications service |
| Posts and users services | File upload/delete commands | File service |
| HTTP services | Structured operational logs | Logs consumer implementation |


## Databases

The development deployment uses PostgreSQL databases owned by service groups:

| Database | Owners in the checked-in seed | Main data |
| --- | --- | --- |
| `config` | Config service | Services and key/value settings |
| `users` | Auth, registration, and users services | Accounts, profiles, confirmation codes, bans, and wallets |
| `posts` | Posts service | Blogs, posts, content, categories, subscriptions, follows, reactions, goals, donations, and income |
| `logs` | Logs service | Structured history logs |
| `comments` | Comments service | Comments and parent lookup index |
| `billing` | Billing service | Robokassa invoices |
| `notifications` | Notifications service | User notifications and seen state |


## API Endpoint groups

| Prefix | Examples |
| --- | --- |
| `/api/v1/registration` | Signup, confirmation, login/email availability |
| `/api/v1/auth` | Login, authorization, user info |
| `/api/v1/users` | Profiles, avatars, user lookup, administration |
| `/api/v1/blogs` | Blog CRUD, categories, follows, subscriptions, donations, goals, income |
| `/api/v1/posts` | Post CRUD, content, attachments, feeds, access, likes, paid access |
| `/api/v1/comments` | Comments by parent and comment counts |
| `/api/v1/notifications` | Inbox, unseen count, mark seen |
| `/api/v1/billing` | Payment links, invoices, callbacks, currency rates |
| `/api/v1/config` | Runtime configuration and administration |
| `/api/v1/logs` | Administrative log queries |

