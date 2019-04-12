package backend

import "testing"

func TestNew(t *testing.T) {
	b, err := New()
	if err != nil {
		t.Errorf("error creating backend: %s", err.Error())
	}

	if b == nil {
		t.Errorf("backend created incorrectly: %+v", *b)
	}
}
