package btrzaws

import (
	"testing"
)

func TestAwsInstanceID(t *testing.T) {
	id, err := GetAwsInstanceID()
	if err != nil {
		t.Fatal("err", err)
	}
	if id != "localhost" && len(id) != 19 {
		t.Fatal("error, bad id:", id)
	}
}
