# Verification

Pounce is verified on the exact repository tree with free, self-hosted tooling.

## Engine

```bash
cd engine
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
go build ./...
```

## Dashboard

```bash
cd dashboard
npm ci
npm run lint
npm run build
npm audit --omit=dev --audit-level=high
```

The browser gate also renders the dashboard in Chromium and checks console/page errors. The screenshot in `docs/screenshots/dashboard.png` was captured from the actual application, not a mockup.

The workflow templates under `.github/workflows/` mirror the same gates and release process. They can be enabled when hosted GitHub Actions are available. Pounce itself remains fully self-hosted and does not require a paid CI, analytics, storage, or deployment service.
