package mediaalert

import "testing"

func TestAlertFor(t *testing.T) {
	tests := []struct {
		name     string
		event    AssetEvent
		wantBody string
		wantSend bool
		wantErr  bool
	}{
		{"ingestion stays quiet", AssetEvent{AssetID: "vid-42", CreatorPhone: "+15550102030", Stage: "ingested"}, "", false, false},
		{"processing start stays quiet", AssetEvent{AssetID: "vid-42", CreatorPhone: "+15550102030", Stage: "processing_started"}, "", false, false},
		{"failed job alerts creator", AssetEvent{AssetID: "vid-42", CreatorPhone: "+15550102030", Stage: "processing_failed", JobID: "job-9", Reason: "audio track missing"}, "Asset vid-42 processing failed in job job-9: audio track missing", true, false},
		{"delivery alerts creator", AssetEvent{AssetID: "vid-42", CreatorPhone: "+15550102030", Stage: "creator_delivered", Title: "Friday stream"}, "Friday stream is processed and delivered. Asset: vid-42", true, false},
		{"failed job needs details", AssetEvent{AssetID: "vid-42", CreatorPhone: "+15550102030", Stage: "processing_failed"}, "", false, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, send, err := AlertFor(test.event)
			if (err != nil) != test.wantErr {
				t.Fatalf("AlertFor() error = %v, wantErr %v", err, test.wantErr)
			}
			if send != test.wantSend || got.Body != test.wantBody {
				t.Fatalf("AlertFor() = (%q, %v), want (%q, %v)", got.Body, send, test.wantBody, test.wantSend)
			}
			if send && got.IdempotencyKey != "asset/vid-42/"+test.event.Stage {
				t.Fatalf("idempotency key = %q", got.IdempotencyKey)
			}
		})
	}
}
