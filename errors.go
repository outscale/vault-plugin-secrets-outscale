package outscalevault

import (
	"errors"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/outscale/osc-sdk-go/v3/pkg/osc"
)

// Translates an error into a logical response if possible, depending on the error type.
// API errors are translated into logical.ErrorResponse, while other errors are returned as Go errors.
func (b *backend) responseFromErr(op string, err error) (*logical.Response, error) {
	if err == nil {
		return nil, nil
	}

	if errors.Is(err, errConfigNotSet) {
		return logical.ErrorResponse(err.Error()), nil
	}

	apiErr := osc.AsErrorResponse(err)
	title := "OSC API request failed"
	if apiErr == nil {
		b.Logger().Error(
			title,
			"operation", op,
			"error", err,
		)

		// If the error is not an API error, return it as a Go error.
		// It will be handled as a Vault Internal Error (500)
		return nil, err
	}

	reqID := *apiErr.ResponseContext.RequestId

	b.Logger().Error(
		title,
		"operation", op,
		"api_error", err,
		"api_request_id", reqID,
	)

	if osc.IsAuthError(err) {
		return logical.ErrorResponse(
			"OSC API authentication failed: %v (request ID: %s); verify the backend authentication configuration",
			err,
			reqID,
		), nil
	}

	return logical.ErrorResponse(
		"%s while %s: %v (request ID: %s)",
		title,
		op,
		err,
		reqID,
	), nil
}
