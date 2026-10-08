package panel

import "testing"

func TestForkUpdaterUsesSelectedRelease(t *testing.T) {
	for _, tag := range []string{"v3.9.0-fourgetu.1", "v3.8.5-fourgetu.1", "v3.9.0", devReleaseTag} {
		ref := tag
		if ref == devReleaseTag {
			ref = "main"
		}
		want := "https://raw.githubusercontent.com/Fourgetu/3x-ui/" + ref + "/update.sh"
		got, err := panelUpdaterURL(tag)
		if err != nil || got != want {
			t.Fatalf("tag=%q URL=%q err=%v, want %q", tag, got, err, want)
		}
	}
	for _, tag := range []string{"", "../main", "v3.9.0/../../main", "v3.9.0;echo test"} {
		if _, err := panelUpdaterURL(tag); err == nil {
			t.Fatalf("accepted invalid release tag %q", tag)
		}
	}
}
