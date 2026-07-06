# 🌳 Estrutura Completa do Projeto OrkaPay

> Gerado em: 01/07/2026
> Legenda: ✅ Implementado | 🏗️ Scaffold (vazio) | 📄 Arquivo

---

```
orkapay/
│
├── 📄 .clinerules/                          [8 regras de arquitetura ✅]
│   ├── 00-core.md
│   ├── 01-architecture.md
│   ├── 02-golang-standards.md
│   ├── 03-nestjs-standards.md
│   ├── 04-testing.md
│   ├── 05-git-workflow.md
│   ├── 06-error-handling.md
│   └── 07-structure-map.md
│
├── 📄 .memory/                              [5 arquivos de contexto ✅]
│   ├── active-context.md
│   ├── progress.md
│   ├── project-brief.md
│   ├── system-patterns.md
│   └── tech-context.md
│
├── 📄 .env.example
├── 📄 .gitignore
├── 📄 AGENTS.md
├── 📄 docker-compose.dev.yml
├── 📄 docker-compose.yml
├── 📄 LICENSE
├── 📄 Makefile
├── 📄 README.md
├── 📄 update-structure.ps1
│
├── 📁 api-gateway/                          [NestJS - Parcialmente implementado]
│   ├── 📄 Dockerfile
│   ├── 📄 nest-cli.json
│   ├── 📄 package.json
│   ├── 📄 tsconfig.build.json
│   ├── 📄 tsconfig.json
│   ├── 📁 src/
│   │   ├── 📄 app.module.ts ✅
│   │   ├── 📄 main.ts ✅
│   │   ├── 📁 common/
│   │   │   ├── 🏗️ config/
│   │   │   ├── 🏗️ decorators/
│   │   │   ├── 🏗️ filters/
│   │   │   ├── 🏗️ guards/
│   │   │   ├── 🏗️ interceptors/
│   │   │   ├── 🏗️ logger/
│   │   │   └── 🏗️ middlewares/
│   │   ├── 📁 infrastructure/
│   │   │   ├── 🏗️ cache/
│   │   │   ├── 🏗️ database/
│   │   │   ├── 🏗️ grpc/
│   │   │   ├── 📁 http/
│   │   │   │   ├── 🏗️ asaas/
│   │   │   │   ├── 🏗️ boleto/
│   │   │   │   ├── 🏗️ celcoin/
│   │   │   │   ├── 🏗️ fitbank/
│   │   │   │   └── 🏗️ starkbank/
│   │   │   └── 📁 messaging/
│   │   │       ├── 🏗️ kafka/
│   │   │       └── 🏗️ rabbitmq/
│   │   ├── 🏗️ migrations/
│   │   └── 📁 modules/
│   │       ├── 📁 boleto/
│   │       │   ├── 📄 boleto.module.ts ✅
│   │       │   ├── 🏗️ controllers/
│   │       │   ├── 🏗️ dto/
│   │       │   ├── 🏗️ entities/
│   │       │   └── 🏗️ services/
│   │       ├── 📁 cashin/
│   │       │   ├── 📄 cashin.module.ts ✅
│   │       │   ├── 🏗️ controllers/
│   │       │   ├── 🏗️ dto/
│   │       │   ├── 🏗️ entities/
│   │       │   └── 🏗️ services/
│   │       ├── 📁 cashout/
│   │       │   ├── 📄 cashout.module.ts ✅
│   │       │   ├── 🏗️ controllers/
│   │       │   ├── 🏗️ dto/
│   │       │   ├── 🏗️ entities/
│   │       │   └── 🏗️ services/
│   │       ├── 📁 health/
│   │       │   └── 📄 health.module.ts ✅
│   │       ├── 🏗️ merchant/
│   │       ├── 📁 payment/
│   │       │   ├── 📄 payment.module.ts ✅
│   │       │   ├── 🏗️ controllers/
│   │       │   ├── 🏗️ dto/
│   │       │   ├── 🏗️ entities/
│   │       │   ├── 🏗️ repositories/
│   │       │   └── 🏗️ services/
│   │       ├── 📁 reconciliation/
│   │       │   ├── 📄 reconciliation.module.ts ✅
│   │       │   ├── 🏗️ controllers/
│   │       │   ├── 🏗️ dto/
│   │       │   ├── 🏗️ entities/
│   │       │   ├── 🏗️ repositories/
│   │       │   └── 🏗️ services/
│   │       ├── 📁 rewards/
│   │       │   ├── 📄 rewards.module.ts ✅
│   │       │   ├── 🏗️ controllers/
│   │       │   ├── 🏗️ dto/
│   │       │   ├── 🏗️ entities/
│   │       │   └── 🏗️ services/
│   │       └── 🏗️ subscription/
│   └── 📁 test/
│       ├── 🏗️ e2e/
│       └── 🏗️ unit/
│
├── 📁 deployments/
│   ├── 📁 github/
│   │   └── 📁 workflows/
│   │       ├── 📄 cd-production.yaml ✅
│   │       ├── 📄 cd-staging.yaml ✅
│   │       └── 📄 ci.yaml ✅
│   └── 📁 helm/
│       ├── 📁 api-gateway/        + 📄 values.yaml ✅
│       ├── 📁 boleto-engine/      + 📄 values.yaml ✅
│       ├── 📁 cashin-engine/      + 📄 values.yaml ✅
│       ├── 📁 cashout-engine/     + 📄 values.yaml ✅
│       ├── 📁 notification/       + 📄 values.yaml ✅
│       ├── 📁 payment-core/       + 📄 values.yaml ✅
│       ├── 📁 reconciliation/     + 📄 values.yaml ✅
│       ├── 📁 rewards-engine/     + 📄 values.yaml ✅
│       └── 📁 smart-retry/        + 📄 values.yaml ✅
│
├── 📁 docs/
│   ├── 📄 estrutura.md ✅
│   ├── 📁 adr/
│   │   ├── 📄 001-event-sourcing.md ✅
│   │   ├── 📄 002-message-broker.md ✅
│   │   ├── 📄 003-smart-retry.md ✅
│   │   ├── 📄 004-baas-providers.md ✅
│   │   ├── 📄 005-cash-in-cash-out.md ✅
│   │   ├── 📄 006-boleto-hybrid-orchestration.md ✅
│   │   └── 📄 007-revenue-share-model.md ✅
│   ├── 📁 api/
│   │   └── 📄 openapi.yaml ✅
│   ├── 📁 business/
│   │   ├── 📄 business-model.md ✅
│   │   ├── 📄 client-onboarding.md ✅
│   │   ├── 📄 financial-projections.md ✅
│   │   ├── 📄 partnership-program.md ✅
│   │   └── 📄 README.md ✅
│   ├── 📁 pricing/
│   │   ├── 📄 competitor-analysis.md ✅
│   │   ├── 📄 plans-comparison.md ✅
│   │   ├── 📄 pricing-table.md ✅
│   │   ├── 📄 README.md ✅
│   │   ├── 📄 revenue-share-policy.md ✅
│   │   └── 📄 roi-calculator.md ✅
│   └── 📁 runbooks/
│       ├── 📄 incident-response.md ✅
│       └── 📄 on-call.md ✅
│
├── 📁 infrastructure/
│   ├── 📁 kubernetes/
│   │   ├── 📁 base/
│   │   │   ├── 📄 configmap.yaml ✅
│   │   │   ├── 📄 namespace.yaml ✅
│   │   │   └── 📄 secrets.yaml ✅
│   │   └── 📁 overlays/
│   │       ├── 🏗️ dev/
│   │       ├── 🏗️ production/
│   │       └── 🏗️ staging/
│   ├── 📁 monitoring/
│   │   ├── 🏗️ alerts/
│   │   ├── 🏗️ dashboards/
│   │   └── 🏗️ logging/
│   └── 📁 terraform/
│       ├── 📁 environments/
│       │   ├── 🏗️ dev/
│       │   ├── 🏗️ production/
│       │   └── 🏗️ staging/
│       └── 📁 modules/
│           ├── 🏗️ database/
│           ├── 🏗️ kubernetes/
│           └── 🏗️ networking/
│
├── 📁 proto/
│   ├── 📄 boleto/boleto.proto ✅
│   ├── 📄 cashin/cashin.proto ✅
│   ├── 📄 cashout/cashout.proto ✅
│   ├── 📄 payment/payment.proto ✅
│   └── 📄 subscription/subscription.proto ✅
│
├── 📁 scripts/
│   ├── 🏗️ deployment/
│   ├── 🏗️ migrations/
│   └── 🏗️ seed/
│
├── 📁 services/
│   │
│   ├── 📁 boleto-engine/                     [🏗️ Scaffold]
│   │   ├── 📄 Dockerfile ✅
│   │   ├── 📄 go.mod ✅
│   │   ├── 📄 go.sum ✅
│   │   ├── 📄 Makefile ✅
│   │   ├── 📁 cmd/
│   │   │   ├── 📄 main.go ✅
│   │   │   ├── 🏗️ api/
│   │   │   └── 🏗️ consumer/kafka/
│   │   ├── 📁 internal/
│   │   │   ├── 📁 adapter/
│   │   │   │   ├── 📁 inbound/
│   │   │   │   │   ├── 🏗️ grpc/
│   │   │   │   │   └── 🏗️ rest/
│   │   │   │   └── 📁 outbound/
│   │   │   │       ├── 🏗️ postgres/
│   │   │   │       └── 🏗️ provider/
│   │   │   ├── 📁 core/
│   │   │   │   ├── 📁 domain/
│   │   │   │   │   ├── 🏗️ entity/
│   │   │   │   │   └── 🏗️ event/
│   │   │   │   └── 📁 port/
│   │   │   │       ├── 🏗️ inbound/
│   │   │   │       └── 🏗️ outbound/
│   │   │   └── 🏗️ service/
│   │   └── 🏗️ pkg/
│   │
│   ├── 📁 cashin-engine/                     [🏗️ Scaffold]
│   │   ├── 📄 Dockerfile ✅
│   │   ├── 📄 go.mod ✅
│   │   ├── 📄 go.sum ✅
│   │   ├── 📄 Makefile ✅
│   │   ├── 📁 cmd/
│   │   │   ├── 📄 main.go ✅
│   │   │   ├── 🏗️ api/
│   │   │   └── 🏗️ consumer/
│   │   ├── 📁 internal/
│   │   │   ├── 📁 adapter/
│   │   │   │   ├── 📁 inbound/
│   │   │   │   │   ├── 🏗️ grpc/
│   │   │   │   │   └── 🏗️ rest/
│   │   │   │   └── 📁 outbound/
│   │   │   │       ├── 🏗️ asaas/
│   │   │   │       └── 🏗️ postgres/
│   │   │   ├── 📁 core/
│   │   │   │   ├── 📁 domain/
│   │   │   │   │   ├── 🏗️ entity/
│   │   │   │   │   └── 🏗️ event/
│   │   │   │   └── 📁 port/
│   │   │   │       ├── 🏗️ inbound/
│   │   │   │       └── 🏗️ outbound/
│   │   │   └── 🏗️ service/
│   │   └── 🏗️ pkg/
│   │
│   ├── 📁 cashout-engine/                    [🏗️ Scaffold]
│   │   ├── 📄 Dockerfile ✅
│   │   ├── 📄 go.mod ✅
│   │   ├── 📄 go.sum ✅
│   │   ├── 📄 Makefile ✅
│   │   ├── 📁 cmd/
│   │   │   ├── 📄 main.go ✅
│   │   │   ├── 🏗️ api/
│   │   │   └── 🏗️ consumer/
│   │   ├── 📁 internal/
│   │   │   ├── 📁 adapter/
│   │   │   │   ├── 📁 inbound/
│   │   │   │   │   ├── 🏗️ grpc/
│   │   │   │   │   └── 🏗️ rest/
│   │   │   │   └── 📁 outbound/
│   │   │   │       ├── 🏗️ postgres/
│   │   │   │       └── 🏗️ starkbank/
│   │   │   ├── 📁 core/
│   │   │   │   ├── 📁 domain/
│   │   │   │   │   ├── 🏗️ entity/
│   │   │   │   │   └── 🏗️ event/
│   │   │   │   └── 📁 port/
│   │   │   │       ├── 🏗️ inbound/
│   │   │   │       └── 🏗️ outbound/
│   │   │   └── 🏗️ service/
│   │   └── 🏗️ pkg/
│   │
│   ├── 📁 notification/                      [🏗️ Scaffold]
│   │   ├── 📄 Dockerfile ✅
│   │   ├── 📄 go.mod ✅
│   │   ├── 📄 go.sum ✅
│   │   ├── 📄 Makefile ✅
│   │   ├── 📁 cmd/
│   │   │   ├── 📄 main.go ✅
│   │   │   └── 🏗️ consumer/
│   │   ├── 📁 internal/
│   │   │   ├── 📁 adapter/
│   │   │   │   ├── 📁 inbound/
│   │   │   │   │   └── 🏗️ consumer/kafka/
│   │   │   │   └── 📁 outbound/
│   │   │   │       ├── 🏗️ email/
│   │   │   │       └── 🏗️ webhook/
│   │   │   ├── 🏗️ core/
│   │   │   └── 🏗️ service/
│   │   └── 🏗️ pkg/
│   │
│   ├── 📁 payment-core/                      [✅ IMPLEMENTADO]
│   │   ├── 📄 Dockerfile ✅
│   │   ├── 📄 go.mod ✅
│   │   ├── 📄 go.sum ✅
│   │   ├── 📄 Makefile ✅
│   │   ├── 📄 coverage, coverage.html, coverage.out ✅
│   │   ├── 📁 cmd/
│   │   │   ├── 📄 main.go ✅
│   │   │   ├── 🏗️ api/
│   │   │   ├── 🏗️ consumer/kafka/
│   │   │   ├── 🏗️ consumer/rabbitmq/
│   │   │   ├── 🏗️ scheduler/
│   │   │   └── 📄 server/main.go ✅
│   │   ├── 📁 internal/
│   │   │   ├── 📁 adapter/
│   │   │   │   ├── 📁 inbound/
│   │   │   │   │   ├── 🏗️ consumer/kafka/
│   │   │   │   │   ├── 🏗️ consumer/rabbitmq/
│   │   │   │   │   ├── 🏗️ grpc/
│   │   │   │   │   └── 📁 rest/
│   │   │   │   │       ├── 🏗️ dto/
│   │   │   │   │       ├── 📁 handler/
│   │   │   │   │       │   ├── 📄 health_handler.go ✅
│   │   │   │   │       │   ├── 📄 mocks_test.go ✅
│   │   │   │   │       │   ├── 📄 payment_handler.go ✅
│   │   │   │   │       │   └── 📄 payment_handler_test.go ✅
│   │   │   │   │       ├── 🏗️ middleware/
│   │   │   │   │       └── 📄 router/router.go ✅
│   │   │   │   └── 📁 outbound/
│   │   │   │       ├── 🏗️ eventstore/
│   │   │   │       ├── 📁 gateway/
│   │   │   │       │   ├── 🏗️ asaas/
│   │   │   │       │   ├── 🏗️ boleto/
│   │   │   │       │   ├── 🏗️ celcoin/
│   │   │   │       │   ├── 🏗️ fitbank/
│   │   │   │       │   └── 🏗️ starkbank/
│   │   │   │       ├── 📁 messaging/
│   │   │   │       │   ├── 📄 kafka/mock_publisher.go ✅
│   │   │   │       │   └── 🏗️ rabbitmq/
│   │   │   │       └── 📁 postgres/
│   │   │   │           ├── 📄 db.go ✅
│   │   │   │           ├── 📄 mapper.go ✅
│   │   │   │           ├── 📄 payment_repository.go ✅
│   │   │   │           ├── 📄 payment_repository_test.go ✅
│   │   │   │           └── 📁 migrations/
│   │   │   │               ├── 📄 001_create_payments_table.up.sql ✅
│   │   │   │               └── 📄 001_create_payments_table.down.sql ✅
│   │   │   ├── 📁 core/
│   │   │   │   ├── 📁 application/
│   │   │   │   │   ├── 📄 cancel_payment.go ✅
│   │   │   │   │   ├── 📄 cancel_payment_test.go ✅
│   │   │   │   │   ├── 📄 create_payment.go ✅
│   │   │   │   │   ├── 📄 create_payment_test.go ✅
│   │   │   │   │   ├── 📄 get_payment.go ✅
│   │   │   │   │   ├── 📄 get_payment_test.go ✅
│   │   │   │   │   ├── 📄 mocks_test.go ✅
│   │   │   │   │   ├── 📄 process_payment.go ✅
│   │   │   │   │   ├── 📄 process_payment_test.go ✅
│   │   │   │   │   ├── 📄 webhook_payment.go ✅
│   │   │   │   │   ├── 📄 webhook_payment_test.go ✅
│   │   │   │   │   └── 🏗️ dto/
│   │   │   │   └── 📁 domain/
│   │   │   │       ├── 📄 errors.go ✅
│   │   │   │       ├── 📄 money.go ✅
│   │   │   │       ├── 📄 money_test.go ✅
│   │   │   │       ├── 📄 payment.go ✅
│   │   │   │       ├── 📄 payment_test.go ✅
│   │   │   │       ├── 📄 pix_key.go ✅
│   │   │   │       ├── 📄 pix_key_test.go ✅
│   │   │   │       ├── 🏗️ entity/
│   │   │   │       └── 📄 event/domain_event.go ✅
│   │   │   ├── 📁 port/
│   │   │   │   ├── 📄 event_publisher.go ✅
│   │   │   │   ├── 📄 payment_gateway.go ✅
│   │   │   │   ├── 📄 payment_getter.go ✅
│   │   │   │   ├── 📄 payment_repository.go ✅
│   │   │   │   ├── 📄 payment_service.go ✅
│   │   │   │   └── 🏗️ outbound/event_store/
│   │   │   ├── 📁 infrastructure/
│   │   │   │   ├── 🏗️ cache/
│   │   │   │   ├── 🏗️ config/
│   │   │   │   ├── 🏗️ database/
│   │   │   │   ├── 🏗️ logger/
│   │   │   │   └── 🏗️ metrics/
│   │   │   └── 📁 service/
│   │   │       ├── 🏗️ boleto/
│   │   │       ├── 🏗️ cashback/
│   │   │       ├── 🏗️ payment/
│   │   │       └── 🏗️ subscription/
│   │   └── 📁 pkg/
│   │       ├── 🏗️ config/
│   │       ├── 🏗️ crypto/
│   │       ├── 🏗️ errors/
│   │       └── 🏗️ utils/
│   │
│   ├── 📁 reconciliation/                    [🏗️ Scaffold]
│   │   ├── 📄 Dockerfile ✅
│   │   ├── 📄 go.mod ✅
│   │   ├── 📄 go.sum ✅
│   │   ├── 📄 Makefile ✅
│   │   ├── 📁 cmd/
│   │   │   ├── 📄 main.go ✅
│   │   │   └── 🏗️ consumer/
│   │   ├── 📁 internal/
│   │   │   ├── 📁 adapter/
│   │   │   │   ├── 📁 inbound/
│   │   │   │   │   └── 🏗️ consumer/kafka/
│   │   │   │   └── 📁 outbound/
│   │   │   │       ├── 🏗️ eventstore/
│   │   │   │       └── 🏗️ postgres/
│   │   │   ├── 📁 core/
│   │   │   │   └── 📁 domain/
│   │   │   │       ├── 🏗️ entity/
│   │   │   │       └── 🏗️ event/
│   │   │   └── 🏗️ service/
│   │   └── 🏗️ pkg/
│   │
│   ├── 📁 rewards-engine/                    [🏗️ Scaffold]
│   │   ├── 📄 Dockerfile ✅
│   │   ├── 📄 go.mod ✅
│   │   ├── 📄 go.sum ✅
│   │   ├── 📄 Makefile ✅
│   │   ├── 📁 cmd/
│   │   │   ├── 📄 main.go ✅
│   │   │   ├── 🏗️ api/
│   │   │   ├── 🏗️ consumer/
│   │   │   └── 🏗️ scheduler/
│   │   ├── 📁 internal/
│   │   │   ├── 📁 adapter/
│   │   │   │   ├── 📁 inbound/
│   │   │   │   │   ├── 🏗️ grpc/
│   │   │   │   │   └── 🏗️ rest/
│   │   │   │   └── 📁 outbound/
│   │   │   │       ├── 🏗️ messaging/kafka/
│   │   │   │       └── 🏗️ postgres/
│   │   │   ├── 📁 core/
│   │   │   │   ├── 📁 domain/
│   │   │   │   │   ├── 🏗️ entity/
│   │   │   │   │   └── 🏗️ event/
│   │   │   │   └── 📁 port/
│   │   │   │       ├── 🏗️ inbound/
│   │   │   │       └── 🏗️ outbound/
│   │   │   └── 🏗️ service/
│   │   └── 🏗️ pkg/
│   │
│   └── 📁 smart-retry/                       [🏗️ Scaffold]
│       ├── 📄 Dockerfile ✅
│       ├── 📄 go.mod ✅
│       ├── 📄 go.sum ✅
│       ├── 📄 Makefile ✅
│       ├── 📁 cmd/
│       │   ├── 📄 main.go ✅
│       │   ├── 🏗️ consumer/
│       │   └── 🏗️ scheduler/
│       ├── 📁 internal/
│       │   ├── 📁 adapter/
│       │   │   ├── 📁 inbound/
│       │   │   │   └── 🏗️ consumer/rabbitmq/
│       │   │   └── 📁 outbound/
│       │   │       ├── 🏗️ messaging/kafka/
│       │   │       └── 🏗️ ml/
│       │   ├── 📁 core/
│       │   │   ├── 🏗️ domain/
│       │   │   └── 🏗️ port/
│       │   └── 📁 service/
│       │       └── 🏗️ retry/
│       └── 🏗️ pkg/
│
└── 📁 shared/
    ├── 📁 baas/
    │   ├── 📁 asaas/
    │   │   ├── 📄 invoice.go ✅
    │   │   └── 📄 pix_auto.go ✅
    │   ├── 📁 boleto/
    │   │   ├── 📄 boleto.go ✅
    │   │   └── 📄 hybrid.go ✅
    │   ├── 🏗️ celcoin/
    │   ├── 🏗️ fitbank/
    │   └── 📁 starkbank/
    │       ├── 📄 batch.go ✅
    │       └── 📄 transfer.go ✅
    ├── 📁 contracts/
    │   ├── 📁 boleto/
    │   │   ├── 📄 boleto_issued.avsc ✅
    │   │   └── 📄 boleto_paid.avsc ✅
    │   ├── 📁 cashin/
    │   │   ├── 📄 invoice_created.avsc ✅
    │   │   └── 📄 invoice_paid.avsc ✅
    │   ├── 📁 cashout/
    │   │   ├── 📄 transfer_created.avsc ✅
    │   │   └── 📄 transfer_settled.avsc ✅
    │   ├── 📁 payment/
    │   │   ├── 📄 payment_created.avsc ✅
    │   │   └── 📄 payment_settled.avsc ✅
    │   └── 📁 rewards/
    │       ├── 📄 cashback_calculated.avsc ✅
    │       ├── 📄 cashback_credited.avsc ✅
    │       └── 📄 revenue_share_distributed.avsc ✅
    └── 🏗️ types/
```

---

## 📊 Resumo por Categoria

| Categoria | Status | Detalhes |
|-----------|--------|----------|
| **🎯 Core (payment-core)** | ✅ **Implementado** | Domínio (Payment, Money, PixKey), Casos de Uso (CRUD + Webhook), Handler REST, Repositório Postgres, Ports, Testes unitários |
| **📚 Documentação** | ✅ **Completo** | ADRs (7), API OpenAPI, Business Model, Pricing, Runbooks |
| **📦 Infra Básica** | ✅ **Pronto** | Helm Charts (9 serviços), GitHub Actions (CI/CD), K8s base, Proto (5), Shared/BAAS (asaas, boleto, starkbank) |
| **📨 Contratos Ávro** | ✅ **Completo** | Schemas para boleto, cashin, cashout, payment, rewards |
| **🌐 API Gateway** | 🏗️ **Parcial** | NestJS scaffold com módulos, apenas `module.ts` implementados |
| **🔧 Demais Serviços** | 🏗️ **Scaffold** | boleto-engine, cashin-engine, cashout-engine, notification, reconciliation, rewards-engine, smart-retry — apenas `cmd/main.go` + estrutura vazia |
| **⏳ Pendentes** | 🏗️ **Vazio** | Terraform modules, Monitoring/Dashboards, Scripts deploy/seed, Gateways (asaas, celcoin, etc.), Consumers Kafka/RabbitMQ, Serviços internos, Infra (cache, config, database, logger, metrics) |
