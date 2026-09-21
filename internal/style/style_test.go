package style

import "testing"

func TestPaletteOn(t *testing.T) {
	p := Palette{on: true}
	cases := map[string]string{
		p.Cyan("x"):       "\x1b[36mx\x1b[0m",
		p.Bold("x"):       "\x1b[1mx\x1b[0m",
		p.Dim("x"):        "\x1b[2mx\x1b[0m",
		p.Green("x"):      "\x1b[32mx\x1b[0m",
		p.Bold(p.Cyan(y)): "\x1b[1m\x1b[36my\x1b[0m\x1b[0m",
	}
	for got, want := range cases {
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

const y = "y"

func TestPaletteOff(t *testing.T) {
	p := Palette{on: false}
	for _, s := range []string{"x", "plain text", "with spaces"} {
		if got := p.Bold(p.Cyan(s)); got != s {
			t.Fatalf("palette off must not wrap text: got %q want %q", got, s)
		}
	}
}

func TestOn(t *testing.T) {
	if !(Palette{on: true}).On() {
		t.Fatal("On() should report true for an enabled palette")
	}
	if (Palette{on: false}).On() {
		t.Fatal("On() should report false for a disabled palette")
	}
}