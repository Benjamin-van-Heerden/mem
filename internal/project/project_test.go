package project

import "testing"

func TestSlugifySeparatesWordsJoinedByPunctuation(t *testing.T) {
	cases := map[string]string{
		"init --template and mem template use/list": "init_template_and_mem_template_use_list",
		"Benjamin van Heerden":                      "benjamin_van_heerden",
		"Upgrade to v1.2: faster sync":              "upgrade_to_v1_2_faster_sync",
		"Fix `mem` crash (again)!":                  "fix_mem_crash_again",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
