# K8s SecretClam Operator & REST API
Проект реализует Kubernetes оператор для кастомного ресурса SecretClam, который управляет дочерними Secrets на основе родительских SecretClam объектов, а также REST API сервер для CRUD операций над этими ресурсами.​

<a href="./cli/README.md">Для CLI инструмента смотрите отдельный README в ./cli >>>>>>></a> 

## Установка и запуск
Установите зависимости и соберите образы одной командой:

```bash
make docker-build
make help  # Полный список команд
```

### Важно: Перед развертыванием создайте JWT секрет для REST API (в namespace default):
```bash
kubectl create secret generic api-jwt-secret \
  --from-literal=JWT_SECRET=super-dev-secret-change-me \
  -n default
```

Или создайте YAML файл api-jwt-secret.yaml:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: api-jwt-secret
  namespace: default
type: Opaque
stringData:
  JWT_SECRET: super-dev-secret-change-me

```
```bash
kubectl apply -f api-jwt-secret.yaml
```

## Полное развертывание в кластере (предполагает Minikube):

```bash
make deploy
```
Эта цель последовательно выполняет: создание namespace k8s-secret-manager-system, развертывание CRD (deploy-crds), RBAC (deploy-rbac), контроллера (deploy-operator) и API сервера (deploy-api).

### Для очистки:

```bash
make undeploy
```
## Локальное тестирование
Запустите контроллер локально для отладки:

### Тестирование без кластера использует envtest:

```bash
make test          # Все тесты (кроме e2e)
make test-operator # Только контроллер
make test-api      # Только API handlers и k8s clients
```


## Tree проекта

```bash
.
├── api
│   ├── openapi.yaml  # контракт для API
│   └── v1alpha1  # Go типы для SecretClaim
│       ├── groupversion_info.go
│       ├── secretclaim_types.go
│       └── zz_generated.deepcopy.go
...
├── cli  # код клиента CLI
│   ├── cmd
│   │   ├── create.go
│   │   ├── delete.go
│   │   ├── get.go
│   │   ├── list.go
│   │   ├── login.go
│   │   ├── root.go
│   │   └── update.go
│   ├── config.json
│   ├── LICENSE
│   ├── main.go
│   └── README.md
├── cmd
│   ├── controller  # Точка входа для контроллера
│   │   └── main.go
│   └── server  # Точка входа для REST API
│       └── main.go
├── config
│   ...
│   ├── custom-rbac  # RBAC для REST API 
│   │   ├── api-server-role.yaml
│   │   ├── api-server-rolebinding.yaml
│   │   ├── api-server-sa.yaml
│   │   ├── leader-election-clusterrole.yaml
│   │   └── leader-election-clusterrolebinding.yaml
│   ...
│   ├── manager # Deployment для контроллера
│   │   ├── kustomization.yaml
│   │   └── manager.yaml
│   ...
│   ├── rbac  # RBAC для контроллера
│   │   ├── kustomization.yaml
│   │   ├── metrics_auth_role_binding.yaml
│   │   ├── metrics_auth_role.yaml
│   │   ├── metrics_reader_role.yaml
│   │   ├── role_binding.yaml
│   │   ├── role.yaml
│   │   └── service_account.yaml
    ...
│   └── server  # Deployment для REST API
│       ├── api-jwt-secret.yaml
│       └── deployment_server.yaml
├── Dockerfile
├── go.mod
├── go.sum
...
├── internal
│   ├── api  # Интерфейс api, сгенерирован из openapi.yaml
│   │   ├── openapi_server.gen.go
│   │   └── server.yaml
│   ├── auth  # Методы для авторизации
│   │   ├── context.go
│   │   ├── helpers.go
│   │   ├── jwt.go
│   │   ├── service.go
│   │   └── types.go
│   ├── cfg
│   │   ├── config.go
│   │   └── users-config.yaml
│   ├── controller  # k8s оператор
│   │   ├── helpers.go
│   │   ├── secretclaim_controller_test.go
│   │   ├── secretclaim_controller.go
│   │   └── suite_test.go
│   ├── handlers  # Хэндлеры для обработки запросов
│   │   ├── auth.go
│   │   ├── error_utils.go
│   │   ├── helpers.go
│   │   ├── secrets_auth_test.go
│   │   ├── secrets.go
│   │   └── server.go
│   ├── k8s  # k8s слой для работы с SecretClaim
│   │   ├── interface.go
│   │   ├── secrets_methods_test.go
│   │   └── secrets_methods.go
│   ├── middleware   # Middleware для авторизации, инъекции ролей и доступов в claim jwt 
│   │   ├── auth_middleware.go
│   │   ├── error_utils.go
│   │   └── helpers.go
│   └── observability  # Middleware и функции для инициализации логгера и трейсера 
│       ├── helpers.go
│       └── init.go
├── ksec  # Бинарный файл CLI 
├── Makefile
├── PROJECT
├── README.md
└── test
    ├── e2e  # TODO
    │   ├── e2e_suite_test.go
    │   └── e2e_test.go
    └── utils
        └── utils.go

```

### REST API

Сервер предоставляет CRUD для SecretClam ресурсов через HTTP endpoints. 

### Аутентификация
JWT Bearer token. Получите токен через CLI (`ksec login`) или endpoint `/auth/login`.

Authorization: Bearer <jwt-token>

### Доступ к API 

**(текущий способ port forward в отдельном терминале)**

```bash
kubectl port-forward svc/api-server 8080:8080  # http://localhost:8080
```

### Endpoints

| Метод | Endpoint | Описание 
|-------|----------|----------|
| POST | `/secrets` | Создать Secret |
| GET | `/secrets` | Список Secrets |
| GET | `/secrets/{name}` | Получить Secret | 
| PUT | `/secrets/{name}` | Обновить Secret | 
| DELETE | `/secrets/{name}` | Удалить Secret | 
| POST | `/user/auth` | Получить JWT |

**OpenAPI**: `api/openapi.yaml` содержит полную спецификацию со схемами

**Примечание**: Namespace передается как query-параметр `?namespace=default`




### Архитектура
- SecretClam CRD: Определяет спецификацию родительского секрета, оператор создает/обновляет дочерние Secrets

- Контроллер: Reconcile loop следит за SecretClam, обеспечивает desired state

- REST API: Отдельный сервер предоставляет HTTP CRUD endpoints над SecretClam ресурсами

- Развертывание: Один Dockerfile строит оба таргета (controller/api-server), Kustomize генерирует manifests
