package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/asecurityteam/transport"
)

type Reporter interface {
	Report(ctx context.Context, payload Output) error
}

// EventReporter is a Reporter implementation which reports events via network to a streaming appliance
type EventReporter struct {
	Endpoint *url.URL
	Client   *http.Client
}

// Reports an event via network to a streaming appliance
func (q *EventReporter) Report(ctx context.Context, payload Output) error {
	body := payload
	rawBody, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, q.Endpoint.String(), bytes.NewReader(rawBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := q.Client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from streaming appliance: %d", res.StatusCode)
	}
	return nil
}

func NewEventReporter(streamApplianceURL *url.URL) (*EventReporter, error) {
	retrier := transport.NewRetrier(
		transport.NewFixedBackoffPolicy(50*time.Millisecond),
		transport.NewLimitedRetryPolicy(3),
		transport.NewStatusCodeRetryPolicy(500, 502, 503),
	)
	base := transport.NewFactory(
		transport.OptionDefaultTransport,
		transport.OptionDisableCompression(true),
		transport.OptionTLSHandshakeTimeout(time.Second),
		transport.OptionMaxIdleConns(100),
	)
	recycler := transport.NewRecycler(
		transport.Chain{retrier}.ApplyFactory(base),
		transport.RecycleOptionTTL(10*time.Minute),
		transport.RecycleOptionTTLJitter(time.Minute),
	)
	httpClient := &http.Client{Transport: recycler}
	eventReporter := EventReporter{
		Client:   httpClient,
		Endpoint: streamApplianceURL,
	}
	return &eventReporter, nil
}
