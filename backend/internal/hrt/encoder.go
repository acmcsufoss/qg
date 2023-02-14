package hrt

import (
	"encoding/json"
	"net/http"
)

// Encoder describes an encoder that encodes or decodes the request and response
// types.
type Encoder interface {
	// Encode encodes the given value into the given writer.
	Encode(http.ResponseWriter, any) error
	// Decode decodes the given value from the given reader.
	Decode(*http.Request, any) error
}

// JSONEncoder is an encoder that encodes and decodes JSON.
var JSONEncoder Encoder = jsonEncoder{}

type jsonEncoder struct{}

func (e jsonEncoder) Encode(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func (e jsonEncoder) Decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// Validator describes a type that can validate itself.
type Validator interface {
	Validate() error
}

// EncoderWithValidator wraps an encoder with one that calls Validate() on the
// value after decoding and before encoding if the value implements Validator.
func EncoderWithValidator(enc Encoder) Encoder {
	return validatorEncoder{enc}
}

type validatorEncoder struct{ enc Encoder }

func (e validatorEncoder) Encode(w http.ResponseWriter, v any) error {
	if validator, ok := v.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	if err := e.enc.Encode(w, v); err != nil {
		return err
	}

	return nil
}

func (e validatorEncoder) Decode(r *http.Request, v any) error {
	if err := e.enc.Decode(r, v); err != nil {
		return err
	}

	if validator, ok := v.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}
