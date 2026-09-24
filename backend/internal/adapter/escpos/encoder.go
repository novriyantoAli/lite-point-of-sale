// Package escpos turns the lines of a Struk into ESC/POS bytes and writes them
// to a thermal printer. Go sends the bytes itself — no CUPS, no `lp`, no JVM
// (ADR-0003, ADR-0017).
//
// The encoder is pure: it takes lines of text and answers bytes, so what a
// receipt's bytes are can be asserted without a device. The printer adapters are
// the only place that touches one.
//
// Lines are written as their UTF-8 bytes, and the printer's code page is never
// selected: the receipt content this app prints is the Indonesian of CONTEXT.md,
// which is ASCII, and choosing a code page would be a guess about a device this
// MVP does not probe. A Produk name outside that set would print as whatever the
// printer's default page makes of those bytes; the paper *layout* is still right,
// because domain/struk wraps by runes rather than bytes.
package escpos

// The ESC/POS control sequences this encoder emits. Named rather than inlined so
// the byte-level tests read as the commands they assert.
const (
	// esc starts every two-byte ESC/POS command.
	esc = 0x1B
	// gs starts every two-byte GS command.
	gs = 0x1D

	// at is ESC @: initialise the printer, clearing its buffer and any mode an
	// earlier job left behind.
	at = '@'

	// alignCommand is ESC a n: set the alignment of everything that follows.
	alignCommand = 'a'
	// alignLeft is the n of a left-aligned receipt. The Struk is laid out with
	// spaces by the domain, so the printer is only ever asked for the left.
	alignLeft = 0

	// boldCommand is ESC E n: turn emphasis on or off.
	boldCommand = 'E'
	// boldOff is the n that turns emphasis off. A printer left bold by an earlier
	// job would print a Struk nobody chose, so Encode says so explicitly.
	boldOff = 0

	// cutCommand is GS V m n: feed the paper to the cutting position and cut it.
	// 0x42 ("feed to cutting position + n dots, then cut") with n = 0 is the one
	// every thermal roll printer this app targets understands.
	cutCommand = 'V'
	cutFeed    = 0x42
	cutDots    = 0
)

// Encode renders the lines of a Struk as the bytes a thermal printer takes:
// initialise, left-align, print every line, then cut the paper.
//
// The domain has already decided the content and the wrapping (ADR-0017,
// keputusan 2), so this is a translation and nothing more: no line is added,
// reordered or trimmed here.
func Encode(lines []string) []byte {
	encoded := []byte{
		esc, at, // initialise
		esc, alignCommand, alignLeft, // left-align
		esc, boldCommand, boldOff, // emphasis off
	}

	for _, line := range lines {
		encoded = append(encoded, line...)
		encoded = append(encoded, '\n')
	}

	return append(encoded, gs, cutCommand, cutFeed, cutDots)
}
