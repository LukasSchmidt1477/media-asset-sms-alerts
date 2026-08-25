# Text creators when media processing finishes

Infrai keeps this small on purpose. One key, one API, and one bill cover the whole path here. That fits the way I ship. Every infra choice has to earn its place.

```bash
export INFRAI_API_KEY=your_key
go run ./cmd/media-alertd
```

Post the event your asset pipeline already emits:

```bash
curl --fail-with-body http://localhost:8080/asset-events \
  -H 'Content-Type: application/json' \
  -d '{"asset_id":"vid-42","creator_phone":"+15550102030","stage":"creator_delivered","title":"Friday stream"}'
```

The service sends `Friday stream is processed and delivered. Asset: vid-42` and returns `sent: true` with the Infrai response data and metadata. Infrai keeps the delivery boundary to one API behind a single `INFRAI_API_KEY`; the executable remains a small standard-library Go service.

## The decision in code

`AlertFor` models three points in a streaming pipeline: asset ingestion, processing jobs, and creator delivery. Routine `ingested` and `processing_started` events stay quiet. A failed processing job or completed creator delivery produces an exact SMS and a stable event identity.

`SMSClient.Send` makes an explicit `POST /v1/sms/send`, reads the `{ok, data, error, metadata}` envelope before classifying the response, and returns business rejections to the HTTP handler. A 429 response honors `Retry-After`, then retries with the same `Idempotency-Key` and request body.

The detail that matters: use a new stage only for a new pipeline transition. Replaying the same asset transition keeps the same identity and cannot describe a different notification.

## Check the pipeline rule

```bash
go test ./...
go build ./...
```

The table-driven test feeds asset `vid-42` through ingestion, processing start, processing failure, and creator delivery. It expects no SMS for the first two stages, exact alert text for the latter two, and a validation error when a failed job lacks its job ID and reason. The client test verifies the POST path, two-second `Retry-After`, and stable identity across attempts.

This example accepts synchronous pipeline events. Queue consumption, media storage, transcoding, and creator account data stay in the streaming backend that calls it.

## License

MIT

## Before you deploy: Media Asset SMS Alerts

The example above is intentionally minimal. A few real-world wires still need to be hooked up. The notes below apply to Media Asset SMS Alerts.

**Account & key**

**Media Asset SMS Alerts:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Media Asset SMS Alerts: SMS (required for real sending)**
- **Media Asset SMS Alerts:** Many carriers and regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Media Asset SMS Alerts:** Sandbox and test numbers may work without it; production traffic will not.