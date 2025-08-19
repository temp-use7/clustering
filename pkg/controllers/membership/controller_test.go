package membership

import (
	"testing"
)

func TestDesiredVotersFuncOverrides(t *testing.T) {
	c := NewController(nil, func() []AliveMember { return nil }).WithDesiredVotersFunc(func() int { return 3 })
	if c.desiredFunc == nil {
		t.Fatalf("desired func not set")
	}
	if v := c.desiredFunc(); v != 3 {
		t.Fatalf("want 3 got %d", v)
	}
}
