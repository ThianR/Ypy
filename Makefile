.PHONY: web-build go-build test
web-build:
	cd web && npm ci && npm run build
go-build:
	go build ./cmd/...
test:
	go test ./...
boundaries:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-boundaries.ps1
verify:
	powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify-all.ps1
