package pm

import (
	"testing"
)

func TestReadRepository(t *testing.T) {
	_, err := NewRepository("../examples/repo/", "tuatzemm/suite/1.0.0")
	if err != nil {
		t.Error(err)
	}
}
