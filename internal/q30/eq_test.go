package q30

import "testing"

func TestPresetID(t *testing.T) {
	if id, ok := PresetID("rock"); !ok || id != 0x11 {
		t.Errorf("PresetID(rock) = %#x, %v", id, ok)
	}
	if _, ok := PresetID("nope"); ok {
		t.Error("PresetID(nope) found")
	}
}

func TestParseBands(t *testing.T) {
	got, err := ParseBands("3, 2,0,0,0,-1.5,2,12")
	if err != nil {
		t.Fatal(err)
	}
	if want := [8]float64{3, 2, 0, 0, 0, -1.5, 2, 12}; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	for _, bad := range []string{"1,2,3", "1,2,3,4,5,6,7,x", "1,2,3,4,5,6,7,13"} {
		if _, err := ParseBands(bad); err == nil {
			t.Errorf("ParseBands(%q) succeeded", bad)
		}
	}
}
