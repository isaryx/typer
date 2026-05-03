package model

import "testing"

func TestParseInputPosition(t *testing.T) {
	cases := []struct {
		in   string
		want InputPlacement
	}{
		{"bottom-left", InputPlacement{V: InputVerticalBottom, H: InputHorizontalLeft}},
		{"BOTTOM-RIGHT", InputPlacement{V: InputVerticalBottom, H: InputHorizontalRight}},
		{"top-center", InputPlacement{V: InputVerticalTop, H: InputHorizontalCenter}},
		{"  bottom-center  ", InputPlacement{V: InputVerticalBottom, H: InputHorizontalCenter}},
		{"tl", InputPlacement{V: InputVerticalTop, H: InputHorizontalLeft}},
		{"BC", InputPlacement{V: InputVerticalBottom, H: InputHorizontalCenter}},
		{" br ", InputPlacement{V: InputVerticalBottom, H: InputHorizontalRight}},
		{"on-top", InputPlacement{V: InputVerticalOnTopBorder, H: InputHorizontalCenter}},
		{"on-bottom", InputPlacement{V: InputVerticalOnBottomBorder, H: InputHorizontalCenter}},
		{"ot", InputPlacement{V: InputVerticalOnTopBorder, H: InputHorizontalCenter}},
		{"ob", InputPlacement{V: InputVerticalOnBottomBorder, H: InputHorizontalCenter}},
		{"on-top-dynamic", InputPlacement{V: InputVerticalOnTopBorderDynamic, H: InputHorizontalCenter}},
		{"on-bottom-dynamic", InputPlacement{V: InputVerticalOnBottomBorderDynamic, H: InputHorizontalCenter}},
		{"otd", InputPlacement{V: InputVerticalOnTopBorderDynamic, H: InputHorizontalCenter}},
		{"obd", InputPlacement{V: InputVerticalOnBottomBorderDynamic, H: InputHorizontalCenter}},
	}
	for _, tc := range cases {
		got, err := ParseInputPosition(tc.in)
		if err != nil {
			t.Fatalf("ParseInputPosition(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseInputPosition(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}

func TestDefaultInputPlacement(t *testing.T) {
	got := DefaultInputPlacement()
	want := InputPlacement{V: InputVerticalOnTopBorderDynamic, H: InputHorizontalCenter}
	if got != want {
		t.Fatalf("DefaultInputPlacement() = %#v, want %#v", got, want)
	}
	if got.CanonicalString() != "on-top-dynamic" {
		t.Fatalf("CanonicalString: got %q", got.CanonicalString())
	}
}

func TestParseInputPositionErrors(t *testing.T) {
	for _, in := range []string{"", "left", "top", "sideways-left", "top-up", "top-centre", "center-left", "top-middle", "bottom-middle", "tb", "lr", "aa", "t"} {
		if _, err := ParseInputPosition(in); err == nil {
			t.Fatalf("ParseInputPosition(%q): expected error", in)
		}
	}
}

func TestInputPlacementCanonicalString(t *testing.T) {
	p := InputPlacement{V: InputVerticalTop, H: InputHorizontalRight}
	canon := p.CanonicalString()
	if canon != "top-right" {
		t.Fatalf("CanonicalString: got %q", canon)
	}
	if got, err := ParseInputPosition(canon); err != nil || got != p {
		t.Fatalf("round-trip: err=%v got=%#v", err, got)
	}
}

func TestBorderPlacementCanonicalRoundTrip(t *testing.T) {
	for _, p := range []InputPlacement{
		{V: InputVerticalOnTopBorder, H: InputHorizontalCenter},
		{V: InputVerticalOnBottomBorder, H: InputHorizontalCenter},
		{V: InputVerticalOnTopBorderDynamic, H: InputHorizontalCenter},
		{V: InputVerticalOnBottomBorderDynamic, H: InputHorizontalCenter},
	} {
		canon := p.CanonicalString()
		got, err := ParseInputPosition(canon)
		if err != nil || got != p {
			t.Fatalf("%#v -> %q -> %#v err=%v", p, canon, got, err)
		}
	}
}
