package archcode

import "testing"

func TestGetArchName(t *testing.T) {
	name := GetArchName()
	if name == "Unknown Architecture" {
		t.Errorf("GetArchName returned Unknown Architecture")
	} else {
		t.Logf("Architecture detected: %s", name)
	}
}
