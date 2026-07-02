package resources

import "testing"

func TestResource(t *testing.T) {
	fs, err := Templates.ReadFile("dist/index.html")
	if err != nil {
		t.Errorf("read file failed, err: %v", err)
	}
	t.Logf("fs: %s", fs)
}
