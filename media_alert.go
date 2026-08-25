package mediaalert

import (
	"context"
	"errors"
	"fmt"
)

type AssetEvent struct {
	AssetID      string `json:"asset_id"`
	CreatorPhone string `json:"creator_phone"`
	Stage        string `json:"stage"`
	JobID        string `json:"job_id,omitempty"`
	Title        string `json:"title,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

type Alert struct {
	To             string
	Body           string
	IdempotencyKey string
}

type Sender interface {
	Send(context.Context, SendRequest) (SendResult, error)
}

type DeliveryNotifier struct {
	SMS Sender
}

func AlertFor(event AssetEvent) (Alert, bool, error) {
	if event.AssetID == "" || event.CreatorPhone == "" {
		return Alert{}, false, errors.New("asset_id and creator_phone are required")
	}

	var body string
	switch event.Stage {
	case "ingested", "processing_started":
		return Alert{}, false, nil
	case "processing_failed":
		if event.JobID == "" || event.Reason == "" {
			return Alert{}, false, errors.New("job_id and reason are required for processing_failed")
		}
		body = fmt.Sprintf("Asset %s processing failed in job %s: %s", event.AssetID, event.JobID, event.Reason)
	case "creator_delivered":
		if event.Title == "" {
			return Alert{}, false, errors.New("title is required for creator_delivered")
		}
		body = fmt.Sprintf("%s is processed and delivered. Asset: %s", event.Title, event.AssetID)
	default:
		return Alert{}, false, fmt.Errorf("unknown asset stage %q", event.Stage)
	}

	return Alert{
		To:             event.CreatorPhone,
		Body:           body,
		IdempotencyKey: "asset/" + event.AssetID + "/" + event.Stage,
	}, true, nil
}

func (n DeliveryNotifier) Handle(ctx context.Context, event AssetEvent) (SendResult, bool, error) {
	alert, send, err := AlertFor(event)
	if err != nil || !send {
		return SendResult{}, false, err
	}
	result, err := n.SMS.Send(ctx, SendRequest{
		To:             alert.To,
		Body:           alert.Body,
		IdempotencyKey: alert.IdempotencyKey,
	})
	return result, true, err
}
