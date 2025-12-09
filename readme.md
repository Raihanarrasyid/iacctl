## 📁 Project Structure

```text
iacctl/
  │
  ├── cmd/
  │   ├── api/
  │   ├── cli/
  │   └── worker/
  │
  ├── internal/
  │   ├── core/
  │   ├── terraform/
  │   ├── queue/
  │   ├── store/
  │   ├── vault/
  │   ├── sandbox/
  │   └── events/
  │
  ├── api/
  │   ├── handler/
  │   ├── middleware/
  │   └── router.go
  │
  ├── configs/
  │   └── config.yaml
  │
  ├── deploy/
  │   └── docker-compose.yml
  │
  ├── scripts/
  │
  ├── web/
  │   └── ...
  │
  ├── pkg/
  │   └── logger/
  │
  ├── test/
  │   ├── integration/
  │   │   └── terraform_apply_test.go
  │   └── mocks/
  │       └── fake_queue.go
  │
  ├── go.mod
  ├── go.sum
  └── README.md
```
