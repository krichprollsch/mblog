package gen

import "testing"

func TestMetadataUnmarshalText(t *testing.T) {
	for _, tc := range []string{
		"", " ",
		"date: 2022-11-14", "title: Hello world",
		"date: 2022-11-14\n", "title: Hello world\n",
		"date: 2022-11-14\ntitle: Hello world",
		"date: 2022-11-14\ntitle: Hello world\n",
		"title: foo: bar",
	} {
		t.Run(tc, func(t *testing.T) {
			var m metadata
			if err := m.UnmarshalText([]byte(tc)); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			t.Logf("%s", m)
		})
	}

	for _, tc := range []string{
		"k", ": v", "k:",
		"date: 14-11-2022",
	} {
		t.Run(tc, func(t *testing.T) {
			var m metadata
			if err := m.UnmarshalText([]byte(tc)); err == nil {
				t.Error("expected error")
			}
			t.Logf("%s", m)
		})
	}
}
