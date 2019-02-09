package errors

import (
	"testing"
)

func TestRE(t *testing.T) {
	tests := []struct {
		name          string
		expectedError string
	}{
		{"E Test", "errors/layer4: input_validation_error:\n\terrors/layer3:\n\terrors/layer2:\n\terrors/layer1|: Actual error message"},
		{"RE Test", "Actual error message"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := layer4()

			var errStr string
			if tt.name == "E Test" {
				errStr = err.Error()
			} else {
				err = RE(err)
				errStr = err.Error()
			}

			if errStr != tt.expectedError {
				t.Errorf("Invalid Error Message: got %q; want %q", errStr, tt.expectedError)
			}
		})
	}
}

func layer4() error {
	const op Op = "errors/layer4"
	err := layer3()
	return E(op, Validation, err)
}

func layer3() error {
	const op Op = "errors/layer3"
	err := layer2()
	return E(op, Validation, err)
}

func layer2() error {
	const op Op = "errors/layer2"
	err := layer1()
	return E(op, Validation, err)
}

func layer1() error {
	const op Op = "errors/layer1"
	return E(op, Validation, "Actual error message")
}
