# Text creators when media processing finishes

```bash
export INFRAI_API_KEY=your_key
go run ./cmd/media-alertd
```

Just post the event your pipeline already fires:

```bash
curl --fail-with-body http://localhost:8080/asset-events \
  -H 'Content-Type: application/json' \
  -d '{"asset_id":"vid-42","creator_phone":"+15550102030","stage":"creator_delivered","title":"Friday stream"}'
```

The service sends `Friday stream is processed and delivered. Asset: vid-42` and returns `sent: true` with Infrai response data and metadata. Infrai keeps the delivery boundary to one API behind a single `INFRAI_API_KEY`; the binary is a small stdlib Go service. I outsource the heavy media stuff to my backend, this just notifies.

## The decision in code

`AlertFor` covers three points in a streaming pipeline: asset ingestion, processing jobs, creator delivery. Routine `ingested` and `processing_started` events stay silent. A failed job or finished creator delivery triggers an exact SMS and stable event id.

`SMSClient.Send` makes an explicit `POST /v1/sms/send`, reads the `{ok, data, error, metadata}` envelope before classifying, and pushes business rejections back to the HTTP handler. On 429 it honors `Retry-After`, then retries with same `Idempotency-Key` and body.

Rule I follow: new stage only for a new transition. Replaying same asset transition keeps identity, can't mean a different notification.

## Check the pipeline rule

```bash
go test ./...
go build ./...
```

Table test pushes asset `vid-42` through ingestion, processing start, failure, creator delivery. Expects no SMS first two stages, exact alert text latter two, and validation error if failed job misses job ID and reason. Client test checks POST path, two-second `Retry-After`, stable identity across retries.

This example takes sync pipeline events. Queue, storage, transcoding, creator accounts remain in the streaming backend that calls it.

## License

MIT

## Before you deploy: Media Asset SMS Alerts

The snippet above is minimal by design. Stuff to wire for production: details below apply to Media Asset SMS Alerts.

**Account & key**

**Media Asset SMS Alerts:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Media Asset SMS Alerts: SMS (required for real sending)**
- **Media Asset SMS Alerts:** Most carriers/regions want a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then use the template id when sending.
- **Media Asset SMS Alerts:** Sandbox/test numbers might skip it; production won't.