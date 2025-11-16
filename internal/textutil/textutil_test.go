package textutil

import "testing"

func TestRemoveConnectors(t *testing.T) {
	in := "Restaurante do Rio de Janeiro e do Mar"
	want := "RESTAURANTE RIO JANEIRO MAR"
	got := RemoveConnectors(in)
	if got != want {
		t.Fatalf("RemoveConnectors(%q) = %q; want %q", in, got, want)
	}
}

func TestRemoveAccents(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Alimentação", "ALIMENTACAO"},
		{"café 123", "CAFE123"},
		{"çãõü", "CAOU"},
	}
	for _, c := range cases {
		got := RemoveAccents(c.in)
		if got != c.want {
			t.Fatalf("RemoveAccents(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}
