# ec-microservices

マイクロサービス・gRPC・Kubernetes・スケーリングを学ぶための簡易 EC バックエンド。

商品閲覧・カート・注文・在庫管理を 4 サービス + Gateway に分割し、gRPC で通信させ、minikube 上で動かす。実務でよく出会う**サービス間呼び出し・在庫競合・スケーリング**を手を動かして体験することを目的とする。

## アーキテクチャ

```
                    [client]
                       │ HTTP/JSON
                       ▼
                ┌─────────────┐
                │   gateway   │
                └──────┬──────┘
                       │ gRPC
        ┌──────────┬───┴────┬───────────┐
        ▼          ▼        ▼           ▼
   ┌─────────┐ ┌──────┐ ┌──────────┐ ┌───────┐
   │ product │ │ cart │ │inventory │ │ order │
   └────┬────┘ └──┬───┘ └────┬─────┘ └───┬───┘
        │         │          │           │
        ▼         ▼          ▼           ▼
  product_db  cart_db  inventory_db  order_db
        └─── 1 つの MySQL に 4 論理 DB ───┘
```

### サービス間通信

| 呼び出し元 | 呼び出し先 | 用途 |
|---|---|---|
| gateway | product / cart / order | HTTP → gRPC 中継 |
| cart | product | 商品存在確認・価格取得 |
| order | product | 商品情報取得 |
| order | inventory | 在庫予約・確定減算 |

## ディレクトリ構成

```
ec-microservices/
├── README.md
├── go.mod
├── Makefile
├── docker-compose.yml          # ローカル開発用(MySQL + 各サービス)
├── proto/                      # gRPC スキーマ定義(契約)
│   ├── product/v1/product.proto
│   ├── cart/v1/cart.proto
│   ├── inventory/v1/inventory.proto
│   └── order/v1/order.proto
├── gen/go/                     # protoc 生成コード(自動生成、編集不可)
├── services/
│   ├── product-service/
│   │   ├── main.go
│   │   ├── server.go           # gRPC ハンドラ
│   │   ├── repo/
│   │   │   ├── repo.go         # Repository インターフェース
│   │   │   ├── memory.go       # in-memory 実装
│   │   │   └── mysql.go        # MySQL 実装
│   │   └── Dockerfile
│   ├── cart-service/           # 同じ構造
│   ├── inventory-service/      # 同じ構造
│   ├── order-service/          # 同じ構造
│   └── gateway/
│       ├── main.go
│       ├── handlers.go         # HTTP ハンドラ + gRPC クライアント
│       └── Dockerfile
├── migrations/                 # golang-migrate 用 SQL
│   ├── product/
│   ├── cart/
│   ├── inventory/
│   └── order/
├── k8s/                        # Kubernetes マニフェスト
│   ├── 00-namespace.yaml
│   ├── 10-mysql.yaml           # StatefulSet + Service + Secret
│   ├── 20-migrations-job.yaml
│   ├── 30-product-service.yaml
│   ├── 31-cart-service.yaml
│   ├── 32-inventory-service.yaml
│   ├── 33-order-service.yaml
│   ├── 40-gateway.yaml
│   └── 50-hpa.yaml
└── scripts/
    ├── seed.sh
    └── load-test.sh
```

## 技術スタック

| 項目 | 採用 | 理由 |
|---|---|---|
| 言語 | Go 1.22+ | gRPC・K8s 周辺ツールと相性最良 |
| 通信 | gRPC(unary) | 型安全、サービス間通信のデファクト |
| 外部 API | HTTP/JSON(gateway のみ) | ブラウザから叩ける |
| DB | MySQL 8 | Database per Service |
| DB アクセス | `database/sql` 直書き | SQL を学ぶ目的のため ORM 不使用 |
| マイグレーション | golang-migrate | デファクト |
| コンテナ | Docker(distroless) | 軽量・本番でも通用 |
| K8s | minikube | ローカル学習用、本物の K8s |
| 負荷試験 | hey | シンプル |

```

## 設計上の意図(なぜこうしたか)

- **Database per Service**: 他サービスのテーブルを直接読まない。必ず gRPC 経由。マイクロサービスの大原則。
- **Repository パターン**: in-memory と MySQL を切り替え可能にすることで、DI と設計の柔軟さを学ぶ。テストも書きやすい。
- **ORM 不使用**: 在庫競合の解決で `SELECT ... FOR UPDATE` を直接書くため。SQL レベルの並行制御を理解する目的。
- **認証・決済を意図的に除外**: 学習の本筋(マイクロサービス・gRPC・K8s・スケーリング)から外れるため。
- **gateway は薄く保つ**: ビジネスロジックを置かない。HTTP ↔ gRPC の変換とエラーコードのマッピングだけ。

---

## 学習ステップ

### Phase 0 — 環境構築

必要なツールをインストールする。

```bash
# Go 1.22+
brew install go

# buf (proto → Go コード生成)
brew install bufbuild/buf/buf

# protoc プラグイン (buf が内部で使う)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Docker Desktop (Docker + Compose 同梱)
# https://www.docker.com/products/docker-desktop/

# minikube + kubectl
brew install minikube kubectl

# grpcurl (gRPC の curl 相当、動作確認に使う)
brew install grpcurl

# hey (負荷試験ツール)
brew install hey
```

確認:
```bash
go version          # go1.22 以上
buf --version
docker --version
minikube version
grpcurl --version
hey
```

---

### Phase 1 — proto 定義と gRPC コード生成

**目標**: gRPC のスキーマ駆動開発を体験する。proto がサービス間の「契約」になることを理解する。

1. `buf.yaml` と `buf.gen.yaml` を作成する
2. `proto/product/v1/product.proto` を書く（message と service の定義）
3. `make proto` で `gen/go/` に Go コードを生成する
4. 生成された `*.pb.go` と `*_grpc.pb.go` を眺めて構造を把握する

```bash
make proto
# gen/go/product/v1/product.pb.go      ← message の struct
# gen/go/product/v1/product_grpc.pb.go ← Client/Server インターフェース
```

**学ぶこと**: proto の message / service / field 型、package・go_package オプション、生成コードの読み方。

---

### Phase 2 — product-service を単体で動かす (in-memory)

**目標**: 一番シンプルなサービスで gRPC サーバーの全体像をつかむ。

1. `services/product-service/repo/repo.go` に Repository インターフェースを定義する
2. `services/product-service/repo/memory.go` に in-memory 実装を書く
3. `services/product-service/server.go` に gRPC ハンドラを実装する
4. `services/product-service/main.go` でサーバーを起動する

```bash
go run ./services/product-service/

# 別ターミナルで動作確認
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext -d '{"name":"T-shirt","price":3000}' \
  localhost:50051 product.v1.ProductService/CreateProduct
grpcurl -plaintext -d '{"id":1}' \
  localhost:50051 product.v1.ProductService/GetProduct
```

**学ぶこと**: gRPC サーバーの起動方法、`grpc.NewServer` / `RegisterXxxServer`、Unary RPC ハンドラのシグネチャ、ステータスコードとエラー返却。

---

### Phase 3 — MySQL + マイグレーション

**目標**: Database per Service パターンと golang-migrate を体験する。

1. `docker-compose.yml` の MySQL だけ起動する
2. `migrations/product/` に up/down の SQL ファイルを作成する
3. `services/product-service/repo/mysql.go` に MySQL 実装を書く
4. `main.go` で `REPO_TYPE=mysql` のとき MySQL リポジトリを選択するよう分岐する

```bash
# MySQL だけ起動
docker compose up -d mysql

# マイグレーション実行
make migrate-product

# MySQL 実装で product-service を起動
REPO_TYPE=mysql DB_DSN="root:password@tcp(localhost:3306)/product_db" \
  go run ./services/product-service/
```

**学ぶこと**: `database/sql` の使い方、マイグレーションファイルの命名規則 (`000001_create_products.up.sql`)、DI による実装の切り替え。

---

### Phase 4 — サービス間 gRPC 呼び出し

**目標**: サービスが他のサービスを gRPC クライアントとして呼ぶ流れを実装する。

#### 4-a: cart-service (product-service を呼ぶ)

1. `services/cart-service/` を実装する
2. `AddItem` の中で product-service の `GetProduct` を呼び、商品の存在確認と価格取得をする
3. 2 サービスを同時に起動して動作確認する

```bash
go run ./services/product-service/ &
PRODUCT_ADDR=localhost:50051 go run ./services/cart-service/ &

grpcurl -plaintext -d '{"user_id":"u1","product_id":1,"quantity":2}' \
  localhost:50052 cart.v1.CartService/AddItem
```

#### 4-b: inventory-service (SELECT FOR UPDATE)

1. `services/inventory-service/` を実装する
2. `ReserveStock` で `BEGIN` → `SELECT quantity FROM stocks WHERE product_id=? FOR UPDATE` → 在庫チェック → `INSERT INTO reservations` → `COMMIT` を書く
3. 並行リクエストで在庫が二重に減らないことを確認する

```bash
# 並行で同じ在庫に予約を投げる
for i in {1..5}; do
  grpcurl -plaintext -d '{"product_id":1,"quantity":10}' \
    localhost:50053 inventory.v1.InventoryService/ReserveStock &
done
wait
```

#### 4-c: order-service (複数サービス呼び出し)

1. `services/order-service/` を実装する
2. `CreateOrder` の中で product-service → inventory-service の順に呼ぶ
3. 在庫不足のときは予約をキャンセルしてエラーを返す（補償トランザクション）

**学ぶこと**: `grpc.NewClient` / `grpc.Dial`、gRPC クライアントの接続管理、エラーコードの伝搬 (`codes.NotFound` など)、`SELECT ... FOR UPDATE` による排他制御。

---

### Phase 5 — gateway (HTTP → gRPC)

**目標**: 外部向け HTTP API がどのように内部 gRPC に変換されるかを実装する。

1. `services/gateway/handlers.go` に HTTP ハンドラを書く
2. 各ハンドラで対応するサービスの gRPC クライアントを呼ぶ
3. gRPC エラーコードを HTTP ステータスコードにマッピングする（`codes.NotFound` → 404 など）

```bash
# 全サービス起動 (docker-compose)
docker compose up

# HTTP で叩けることを確認
curl -s localhost:8080/products | jq
curl -s -X POST localhost:8080/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Hoodie","price":8000}' | jq
```

**学ぶこと**: `net/http` のルーティング（Go 1.22 の `{id}` パス変数）、gRPC ステータス → HTTP ステータスのマッピング、gateway パターンの責務の境界。

---

### Phase 6 — Docker コンテナ化

**目標**: 各サービスを distroless コンテナにパッケージし、docker-compose でフル起動する。

1. 各サービスに `Dockerfile` を書く（マルチステージビルド）
2. `docker-compose.yml` に全サービスを定義する
3. ヘルスチェックと依存順序（`depends_on`）を設定する

```bash
make build-images
docker compose up

# エンドツーエンドで動くことを確認
./scripts/seed.sh
curl localhost:8080/products | jq
```

**学ぶこと**: マルチステージビルド、distroless イメージ、docker-compose の `depends_on` と `healthcheck`、環境変数によるサービス設定。

---

### Phase 7 — Kubernetes (minikube)

**目標**: 本物の K8s 上で動かし、マニフェストの各リソースを理解する。

```bash
minikube start --cpus=4 --memory=6g

# イメージを minikube に読み込む
make minikube-load

# マニフェストを順に適用
kubectl apply -f k8s/00-namespace.yaml
kubectl apply -f k8s/10-mysql.yaml
kubectl wait --for=condition=Ready pod -l app=mysql -n ec --timeout=120s
kubectl apply -f k8s/20-migrations-job.yaml
kubectl wait --for=condition=complete job -l app=migrations -n ec --timeout=60s
kubectl apply -f k8s/30-product-service.yaml
kubectl apply -f k8s/31-cart-service.yaml
kubectl apply -f k8s/32-inventory-service.yaml
kubectl apply -f k8s/33-order-service.yaml
kubectl apply -f k8s/40-gateway.yaml

# 状態確認
kubectl get pods -n ec
kubectl get svc -n ec
```

各マニフェストで学ぶこと:
| ファイル | 学ぶリソース |
|---|---|
| `10-mysql.yaml` | StatefulSet, PersistentVolumeClaim, Secret |
| `20-migrations-job.yaml` | Job (一度だけ実行) |
| `30-*.yaml` | Deployment, Service (ClusterIP), ConfigMap |
| `40-gateway.yaml` | Service (NodePort or LoadBalancer) |
| `50-hpa.yaml` | HorizontalPodAutoscaler |

```bash
# gateway に疎通確認
minikube service gateway -n ec --url
curl <上記URL>/products | jq
```

---

### Phase 8 — スケーリング体験

**目標**: HPA が CPU 負荷に応じてレプリカを増やす様子を観察する。

```bash
# HPA を有効化 (metrics-server が必要)
minikube addons enable metrics-server
kubectl apply -f k8s/50-hpa.yaml

# 別ターミナルで Pod 数を監視
watch kubectl get pods -n ec

# 負荷をかける
hey -n 10000 -c 50 $(minikube service gateway -n ec --url)/products

# HPA の状態確認
kubectl get hpa -n ec
kubectl describe hpa gateway -n ec
```

**学ぶこと**: HPA のスケールアップ/ダウンの閾値と遅延、`resource.requests` の重要性（HPA は requests に対する使用率を見る）、gRPC のサービスディスカバリと負荷分散（ClusterIP + kube-proxy）。

---

### 完了後の発展課題

| テーマ | 内容 |
|---|---|
| 分散トレーシング | OpenTelemetry + Jaeger でリクエストの流れを可視化 |
| サーキットブレーカー | 在庫サービスが落ちたとき注文がどう振る舞うか |
| gRPC streaming | 在庫変動をリアルタイム通知する Server-Streaming RPC |
| Saga パターン | 補償トランザクションを Saga として体系的に実装 |
| サービスメッシュ | Istio/Linkerd で mTLS・トラフィック制御を追加 |