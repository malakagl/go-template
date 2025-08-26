
### Build an API server implementing our OpenAPI spec for food ordering API in Go.

Go version used: 1.25

API Spec Published at https://orderfoodonline.deno.dev/public/openapi.yaml

### How to run

```
make run
```

### Run unit tests

```
make test
```

### Run integration tests

```
make run-it
```

### Useful commands

```
# install dev dependencies
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# to create new migration file
migrate create -ext sql -dir db/migrations -seq init_db

# to start the database
docker compose up -d postgres

# to stop the database
docker compose down postgres 

# apply migations
migrate -path db/migrations -database "postgres://user:password@localhost:5432/test?sslmode=disable" up

# remove migrations
migrate -path db/migrations -database "postgres://user:password@localhost:5432/test?sslmode=disable" down

# install lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 <- fixed with latest version

# check coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# deploy
minikube start --memory 10240 --cpus 4
minikube addons enable ingress
minikube mount ./promocodes:/mnt/promocodes
minikube mount ./db:/mnt/db
docker build -f ./docker/Dockerfile -t kart-challenge:latest .
minikube image load kart-challenge:latest
kubectl apply -k ./deployment/k8s/
minikube tunnel

# mount config and data
kubectl apply -k ./deployment/k8s/

# verify
kubectl get deployments -A
minikube service kart-challenge --url -n kart-challenge
minikube service kart-challenge --url -n postgres
kubectl get pods -n kart-challenge
kubectl logs kart-challenge-5d8ccdfcdd-qh8cb -n kart-challenge -f
kubectl exec -it -n kart-challenge kart-challenge-6d69f6b49-hn5fb -- /bin/sh

# update
kubectl rollout restart deployment/kart-challenge -n kart-challenge

# clean up
kubectl delete service kart-challenge -n kart-challenge
kubectl delete service postgres -n kart-challenge
kubectl delete deployment kart-challenge -n kart-challenge
kubectl delete deployment postgres -n kart-challenge
kubectl delete -n kart-challenge persistentvolumeclaim postgres-pvc
minikube stop
minikube delete --all
```

### Notes
- Took around 25 minutes to load all coupon codes to the database.
- To read gz files via code and scan it took around 15 seconds.

### Benchmark coupon validation

```
go test -bench=. -benchmem ./internal/couponcode/...
```
Using compressed coupon code files.

```
% go test -bench=. -benchmem ./internal/couponcode/...
goos: darwin
goarch: amd64
pkg: github.com/malakagl/kart-challenge/internal/couponcode
cpu: Intel(R) Core(TM) i5-8500 CPU @ 3.00GHz
BenchmarkValidateCouponCode_Valid-6 1 13770662525 ns/op 6093944 B/op 289099 allocs/op
BenchmarkValidateCouponCode_Invalid-6 1 13473384996 ns/op 6129808 B/op 290891 allocs/op
PASS
ok github.com/malakagl/kart-challenge/internal/couponcode 27.598s
```

| Benchmark    | Iterations | Time per op | Allocated Bytes | Allocations |
| ------------ | ---------- | ----------- | --------------- | ----------- |
| Valid code   | 1          | ~13.8s     | ~6 MB          | 289k        |
| Invalid code | 1          | ~13.5s     | ~6.1 MB        | 291k        |

Using decompressed coupon code files.

```
% go test -bench=. -benchmem ./internal/couponcode/...
goos: darwin
goarch: amd64
pkg: github.com/malakagl/kart-challenge/internal/couponcode
cpu: Intel(R) Core(TM) i5-8500 CPU @ 3.00GHz
BenchmarkValidateCouponCode_Valid-6                    1        4274126382 ns/op           14904 B/op         31 allocs/op
BenchmarkValidateCouponCode_Invalid-6                  1        4200461553 ns/op           14376 B/op         29 allocs/op
PASS
ok      github.com/malakagl/kart-challenge/internal/couponcode  8.849s
```
| Benchmark    | Iterations | Time per op | Allocated Bytes | Allocations |
| ------------ | ---------- |-------------|-----------------| ----------- |
| Valid code   | 1          | ~4.27s      | ~14.9 KB        | 31          |
| Invalid code | 1          | ~4.20s      | ~14.3 KB        | 29          |

### Strategy

Spawn a goroutine to decompress files on server start up.
This will take around 1-2 minutes.
Meanwhile, serve any requests using compressed files.
These requests will take around 30 seconds to process. 
But it should be still acceptable given this is a deployment window.
Once decompress of files completed started using decompressed files to validate coupon code.
Also implemented a cache to store a limited number of validated coupon codes.
This way if a malicious user tries a same coupon code multiple times it will not use a lot of server resources.

Another solution is to load all the coupon codes to postgres database.
Initial loading will take a bit of time. But coupon code validation will be much faster.

### Tasks
- [x] Implement the API server
- [x] Implement the OpenAPI spec
- [x] Implement the unit tests
- [x] Implement promo code functionality
- [x] Implement the database
- [x] Implement the database migrations
- [x] Implement logging
- [ ] Implement the error handling
- [x] Implement the configuration management
- [x] Implement the integration tests
- [x] Implement the Dockerfile
- [x] Implement the Makefile
- [x] Implement the README.md
- [ ] Implement the CI/CD pipeline
- [x] Implement the GitHub Actions workflow
- [x] Implement the GitHub Pull Requests
- [ ] Implement money package for handling money
- [x] Implement the caching
- [x] Implement the rate limiting
- [ ] Implement the security
- [ ] Implement the monitoring
- [x] Implement the tracing
- [x] Create private schema for postgres
- [x] Add lint
- [ ] Add idempotency to order create API