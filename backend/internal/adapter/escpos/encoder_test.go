package escpos

import (
	"bytes"
	"testing"
)

// The exact bytes of each command. Asserting the bytes rather than "the encoder
// ran" is the point: an ESC/POS printer either gets the right byte sequence or
// prints garbage (ADR-0017, keputusan 2).
func TestCommandsAreTheBytesAPrinterUnderstands(t *testing.T) {
	tests := []struct {
		name string
		got  []byte
		want []byte
	}{
		{name: "initialise", got: Init(), want: []byte{0x1B, 0x40}},
		{name: "align left", got: Align(Left), want: []byte{0x1B, 0x61, 0x00}},
		{name: "align center", got: Align(Center), want: []byte{0x1B, 0x61, 0x01}},
		{name: "align right", got: Align(Right), want: []byte{0x1B, 0x61, 0x02}},
		{name: "bold on", got: Bold(true), want: []byte{0x1B, 0x45, 0x01}},
		{name: "bold off", got: Bold(false), want: []byte{0x1B, 0x45, 0x00}},
		{name: "cut", got: Cut(), want: []byte{0x1D, 0x56, 0x42, 0x00}},
	}

	for _, test := range tests {
		if !bytes.Equal(test.got, test.want) {
			t.Errorf("%s: got % X, want % X", test.name, test.got, test.want)
		}
	}
}

func TestEncodeWrapsTheLinesInInitAndCut(t *testing.T) {
	lines := []string{"Nomor Struk: 12", "Total   Rp 54.000"}

	got := Encode(lines)

	want := []byte{0x1B, 0x40}            // initialise
	want = append(want, 0x1B, 0x61, 0x00) // left align
	want = append(want, 0x1B, 0x45, 0x00) // emphasis off
	for _, line := range lines {
		want = append(want, []byte(line)...)
		want = append(want, '\n')
	}
	want = append(want, 0x1D, 0x56, 0x42, 0x00) // cut

	if !bytes.Equal(got, want) {
		t.Errorf("Encode: got % X, want % X", got, want)
	}
}

func TestEncodePrintsTheLinesInOrderAndInFull(t *testing.T) {
	lines := []string{"Toko Kopi Purnama", "", "Nomor Struk: 12", "Kembalian   Rp 6.000"}

	encoded := Encode(lines)

	// Each line reaches the printer exactly as the domain wrote it — including the
	// blank one, which is what separates the template block from the sale.
	for _, line := range lines {
		if !bytes.Contains(encoded, []byte(line+"\n")) {
			t.Errorf("Encode % X: want the line %q, newline and all", encoded, line)
		}
	}

	// And in the order they were given, so the Struk reads top to bottom.
	previous := -1
	for _, line := range lines {
		at := bytes.Index(encoded, []byte(line+"\n"))
		if at <= previous {
			t.Errorf("line %q: got offset %d, want it after %d", line, at, previous)
		}
		previous = at
	}
}

func TestEncodeOfNoLinesIsStillAValidJob(t *testing.T) {
	encoded := Encode(nil)

	// Initialise, align, emphasis off, cut — a printer that gets this prints
	// nothing and is left ready for the next Struk.
	if !bytes.HasPrefix(encoded, Init()) {
		t.Errorf("Encode(nil): got % X, want it to start with Init()", encoded)
	}
	if !bytes.HasSuffix(encoded, Cut()) {
		t.Errorf("Encode(nil): got % X, want it to end with Cut()", encoded)
	}
}
