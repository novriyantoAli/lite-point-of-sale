// Package escpos turns the lines of a Struk into ESC/POS bytes and writes them
// to a thermal printer. Go sends the bytes itself — no CUPS, no `lp`, no JVM
// (ADR-0003, ADR-0017).
//
// The encoder is pure: it takes lines of text and answers bytes, so what a
// receipt's bytes are can be asserted without a device. The printer adapters are
// the only place that touches one.
package escpos

// The ESC/POS control sequences this encoder uses. Named rather than inlined so
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
	boldOff     = 0
	boldOn      = 1

	// cutCommand is GS V m n: feed the paper to the cutting position and cut it.
	// 0x42 ("feed to cutting position + n dots, then cut") with n = 0 is the one
	// every thermal roll printer this app targets understands.
	cutCommand = 'V'
	cutFeed    = 0x42
	cutDots    = 0
)

// Init returns the bytes that initialise a printer.
func Init() []byte {
	return []byte{esc, at}
}

// Alignment is one of the three ESC/POS alignment modes.
type Alignment byte

const (
	// Left is the alignment of a Struk: its layout is spaced out by the domain,
	// not centred by the printer.
	Left Alignment = 0
	// Center and Right are the other two modes the command supports.
	Center Alignment = 1
	Right  Alignment = 2
)

// Align returns the bytes that set the alignment of what follows.
func Align(alignment Alignment) []byte {
	return []byte{esc, alignCommand, byte(alignment)}
}

// Bold returns the bytes that turn emphasis on or off.
func Bold(on bool) []byte {
	emphasis := byte(boldOff)
	if on {
		emphasis = boldOn
	}

	return []byte{esc, boldCommand, emphasis}
}

// Cut returns the bytes that feed the paper to the cutting position and cut it.
func Cut() []byte {
	return []byte{gs, cutCommand, cutFeed, cutDots}
}

// Encode renders the lines of a Struk as the bytes a thermal printer takes:
// initialise, left-align, print every line, then cut the paper.
//
// The domain has already decided the content and the wrapping (ADR-0017,
// keputusan 2), so this is a translation and nothing more: no line is added,
// reordered or trimmed here. Emphasis is explicitly turned off because a
// printer left bold by an earlier job would print a Struk nobody chose.
func Encode(lines []string) []byte {
	encoded := Init()
	encoded = append(encoded, Align(Left)...)
	encoded = append(encoded, Bold(false)...)

	for _, line := range lines {
		encoded = append(encoded, line...)
		encoded = append(encoded, '\n')
	}

	return append(encoded, Cut()...)
}
