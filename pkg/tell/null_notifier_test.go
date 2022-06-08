package tell

import (
	"testing"
)

func TestNullNotifier(t *testing.T) {
	event := NullNotifier.Start("the.truth.is.out.there", nil)
	event.Finish()
}
