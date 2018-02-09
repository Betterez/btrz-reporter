package reporter

import (
	"testing"
)

func TestMemLoader(t *testing.T) {
	_, err := LoadMemoryValue()
	//	data, err := LoadMemoryValue()
	if err != nil {
		t.Fatalf("error %v", err)
	}

}
func TestMemLoaderValue(t *testing.T) {
	data, err := LoadMemoryValue()
	if err != nil {
		t.Fatalf("error %v", err)
	}
	if data.GetFreeMemory() == 0 {
		t.Fatal("Free memory is 0")
	}
	if data.GetTotalMemory() == 0 {
		t.Fatal("Total memory is 0")
	}
	if data.GetUsedMemory() == 0 {
		t.Fatal("used memory is 0")
	}
}

func TestOSVersion(t *testing.T) {
	version, err := loadOSVersion()
	if err != nil {
		t.Fatalf("error %v", err)
	}
	if (version != 14.04) && (version != 16.04) {
		t.Fatalf("bad version returns, %f", version)
	}
}
