package struk

import (
	"strings"
	"testing"
	"unicode/utf8"

	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
)

// kopiAndTeh is the Penjualan most tests print: two Items, paid in Tunai with a
// Kembalian.
func kopiAndTeh() domainpenjualan.Sale {
	return domainpenjualan.Sale{
		ReceiptNumber: 12,
		CreatedAt:     "2026-09-23 10:00:00",
		CashierID:     1,
		CashierName:   "admin",
		Total:         2*18000 + 3*6000,
		Items: []domainpenjualan.Item{
			{ProductID: 1, Name: "Kopi Susu", Price: 18000, Quantity: 2},
			{ProductID: 2, Name: "Teh Botol", Price: 6000, Quantity: 3},
		},
		Payment: domainpenjualan.Payment{Method: domainpenjualan.PaymentCash, Amount: 60000, Change: 6000},
	}
}

// indexOf is the position of the first line that equals want, or -1.
func indexOf(lines []string, want string) int {
	for i, line := range lines {
		if line == want {
			return i
		}
	}

	return -1
}

// indexOfPrefix is the position of the first line that starts with prefix.
func indexOfPrefix(lines []string, prefix string) int {
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return i
		}
	}

	return -1
}

// assertRightAligned checks a money line: it starts with left, ends with right,
// and is exactly as wide as the paper. What it deliberately does not assert is
// how many spaces sit between the two — that is layout, not content.
func assertRightAligned(t *testing.T, line, left, right string, columns int) {
	t.Helper()

	if !strings.HasPrefix(line, left) {
		t.Errorf("line %q: want it to start with %q", line, left)
	}
	if !strings.HasSuffix(line, right) {
		t.Errorf("line %q: want it to end with %q", line, right)
	}
	if got := utf8.RuneCountInString(line); got != columns {
		t.Errorf("line %q: got %d columns, want %d", line, got, columns)
	}
}

func TestColumnsDerivesFromThePaperWidth(t *testing.T) {
	tests := []struct {
		paperWidth int64
		want       int
	}{
		{paperWidth: 58, want: 32},
		{paperWidth: 80, want: 48},
		// The stored Pengaturan only ever holds 58 or 80; anything else falls back
		// to the wider roll rather than wrapping to a width nobody chose.
		{paperWidth: 0, want: 48},
		{paperWidth: 210, want: 48},
	}

	for _, test := range tests {
		if got := Columns(test.paperWidth); got != test.want {
			t.Errorf("Columns(%d): got %d, want %d", test.paperWidth, got, test.want)
		}
	}
}

func TestComposeFollowsTheStrukOrder(t *testing.T) {
	sale := kopiAndTeh()
	template := Template{Header: "Toko Kopi Purnama\nJl. Melati 1", Footer: "Terima kasih", PaperWidth: 80}

	lines := Compose(sale, template)

	// The order CONTEXT.md lists (ADR-0017, keputusan 4): header block, Nomor
	// Struk, waktu, Kasir, Item, total, metode, jumlah bayar, Kembalian, footer.
	wantOrder := []string{
		"Toko Kopi Purnama",
		"Jl. Melati 1",
		"Nomor Struk: 12",
		"2026-09-23 10:00:00",
		"Kasir: admin",
		"Kopi Susu",
		"Teh Botol",
		"Total",
		"Bayar (Tunai)",
		"Kembalian",
		"Terima kasih",
	}

	previous := -1
	for _, want := range wantOrder {
		at := indexOfPrefix(lines, want)
		if at < 0 {
			t.Fatalf("lines %q: want a line starting with %q", lines, want)
		}
		if at <= previous {
			t.Errorf("line %q: got position %d, want it after position %d", want, at, previous)
		}
		previous = at
	}
}

func TestComposeRendersTheItemAndTheMoney(t *testing.T) {
	lines := Compose(kopiAndTeh(), Template{PaperWidth: 80})

	if !contains(lines, "Kopi Susu") || !contains(lines, "Teh Botol") {
		t.Fatalf("lines %q: want both Item names", lines)
	}

	kopi := lineWithPrefix(lines, "  2 x ")
	assertRightAligned(t, kopi, "  2 x Rp 18.000", "Rp 36.000", 48)
	teh := lineWithPrefix(lines, "  3 x ")
	assertRightAligned(t, teh, "  3 x Rp 6.000", "Rp 18.000", 48)

	assertRightAligned(t, lineWithPrefix(lines, "Total"), "Total", "Rp 54.000", 48)
	assertRightAligned(t, lineWithPrefix(lines, "Bayar (Tunai)"), "Bayar (Tunai)", "Rp 60.000", 48)
	assertRightAligned(t, lineWithPrefix(lines, "Kembalian"), "Kembalian", "Rp 6.000", 48)
}

func TestComposePrintsTheKembalianOnlyForTunai(t *testing.T) {
	tests := []struct {
		name     string
		method   domainpenjualan.PaymentMethod
		wantLine bool
	}{
		{name: "Tunai", method: domainpenjualan.PaymentCash, wantLine: true},
		{name: "QRIS", method: domainpenjualan.PaymentQRIS, wantLine: false},
		{name: "Debit", method: domainpenjualan.PaymentDebit, wantLine: false},
		{name: "Transfer", method: domainpenjualan.PaymentTransfer, wantLine: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sale := kopiAndTeh()
			sale.Payment = domainpenjualan.Payment{Method: test.method, Amount: sale.Total}

			lines := Compose(sale, Template{PaperWidth: 80})

			if got := indexOfPrefix(lines, "Kembalian"); (got >= 0) != test.wantLine {
				t.Errorf("Kembalian line for %s: got %v, want %v (lines %q)",
					test.name, got >= 0, test.wantLine, lines)
			}

			// The method itself is printed whichever it is.
			if indexOfPrefix(lines, "Bayar ("+methodLabel(test.method)+")") < 0 {
				t.Errorf("lines %q: want a line naming the method %s", lines, test.name)
			}
		})
	}
}

func TestComposeWrapsLongLinesInsteadOfCuttingThem(t *testing.T) {
	name := "Kopi Susu Gula Aren Spesial Untuk Pelanggan Setia Toko"
	sale := domainpenjualan.Sale{
		ReceiptNumber: 1,
		CreatedAt:     "2026-09-23 10:00:00",
		CashierName:   "admin",
		Total:         18000,
		Items:         []domainpenjualan.Item{{ProductID: 1, Name: name, Price: 18000, Quantity: 1}},
		Payment:       domainpenjualan.Payment{Method: domainpenjualan.PaymentCash, Amount: 18000},
	}

	for _, columns := range []int{32, 48} {
		lines := Compose(sale, Template{PaperWidth: paperWidthOf(columns)})

		for _, line := range lines {
			if got := utf8.RuneCountInString(line); got > columns {
				t.Errorf("at %d columns: line %q is %d wide", columns, line, got)
			}
		}

		// Nothing is dropped: the wrapped name lines join back into the name.
		wrapped := lines[indexOf(lines, wrapText(name, columns)[0]):indexOfPrefix(lines, "  1 x ")]
		if got := strings.Join(wrapped, " "); got != name {
			t.Errorf("at %d columns: wrapped name %q, want %q", columns, got, name)
		}
	}
}

func TestComposeWrapsTemplateBlocksToThePaperWidth(t *testing.T) {
	long := "Toko Kopi Purnama Jl. Melati Nomor Satu Kelurahan Sukamaju"
	sale := kopiAndTeh()

	for _, columns := range []int{32, 48} {
		lines := Compose(sale, Template{Header: long, Footer: long, PaperWidth: paperWidthOf(columns)})

		for _, line := range lines {
			if got := utf8.RuneCountInString(line); got > columns {
				t.Errorf("at %d columns: line %q is %d wide", columns, line, got)
			}
		}
	}

	// A short line is printed verbatim, spacing and all.
	verbatim := Compose(sale, Template{Header: "  Toko  Kopi  ", PaperWidth: 80})
	if !contains(verbatim, "  Toko  Kopi  ") {
		t.Errorf("short header: got %q, want it printed verbatim", verbatim)
	}
}

func TestComposeMovesMoneyToItsOwnLineWhenItDoesNotFit(t *testing.T) {
	// At 32 columns, "  10 x Rp 1.250.000" plus "Rp 12.500.000" is wider than the
	// roll, so the subtotal moves to a line of its own rather than being dropped.
	sale := domainpenjualan.Sale{
		ReceiptNumber: 1,
		CreatedAt:     "2026-09-23 10:00:00",
		CashierName:   "admin",
		Total:         12500000,
		Items: []domainpenjualan.Item{
			{ProductID: 1, Name: "Kopi", Price: 1250000, Quantity: 10},
		},
		Payment: domainpenjualan.Payment{Method: domainpenjualan.PaymentCash, Amount: 12500000},
	}

	lines := Compose(sale, Template{PaperWidth: 58})

	detail := lineWithPrefix(lines, "  10 x Rp 1.250.000")
	if detail != "  10 x Rp 1.250.000" {
		t.Errorf("detail line: got %q, want the money to have moved off it", detail)
	}
	assertRightAligned(t, lines[indexOf(lines, detail)+1], "", "Rp 12.500.000", 32)
}

func TestComposeLeavesAnEmptyTemplateOut(t *testing.T) {
	lines := Compose(kopiAndTeh(), Template{PaperWidth: 80})

	// An empty block is no lines at all — the first line is the Nomor Struk, not
	// a blank the Admin never typed.
	if lines[0] != "Nomor Struk: 12" {
		t.Errorf("first line: got %q, want the Nomor Struk", lines[0])
	}
	if !strings.HasPrefix(lines[len(lines)-1], "Kembalian") {
		t.Errorf("last line: got %q, want the Kembalian", lines[len(lines)-1])
	}
}

func TestComposeKeepsABlankLineInsideATemplateBlock(t *testing.T) {
	lines := Compose(kopiAndTeh(), Template{Header: "Toko Kopi\n\nJl. Melati 1", PaperWidth: 80})

	if indexOf(lines, "Toko Kopi")+1 != indexOf(lines, "Jl. Melati 1")-1 {
		t.Errorf("lines %q: want the blank line the Admin typed kept", lines)
	}
}

func TestFormatMoneyGroupsWholeRupiah(t *testing.T) {
	tests := []struct {
		amount int64
		want   string
	}{
		{amount: 0, want: "Rp 0"},
		{amount: 900, want: "Rp 900"},
		{amount: 18000, want: "Rp 18.000"},
		{amount: 1250000, want: "Rp 1.250.000"},
		{amount: 1250000000, want: "Rp 1.250.000.000"},
	}

	for _, test := range tests {
		if got := formatMoney(test.amount); got != test.want {
			t.Errorf("formatMoney(%d): got %q, want %q", test.amount, got, test.want)
		}
	}
}

func TestMethodLabelWritesTheDomainTerm(t *testing.T) {
	tests := []struct {
		method domainpenjualan.PaymentMethod
		want   string
	}{
		{method: domainpenjualan.PaymentCash, want: "Tunai"},
		{method: domainpenjualan.PaymentQRIS, want: "QRIS"},
		{method: domainpenjualan.PaymentDebit, want: "Debit"},
		{method: domainpenjualan.PaymentTransfer, want: "Transfer"},
	}

	for _, test := range tests {
		if got := methodLabel(test.method); got != test.want {
			t.Errorf("methodLabel(%q): got %q, want %q", test.method, got, test.want)
		}
	}
}

func contains(lines []string, want string) bool {
	return indexOf(lines, want) >= 0
}

func lineWithPrefix(lines []string, prefix string) string {
	at := indexOfPrefix(lines, prefix)
	if at < 0 {
		return ""
	}

	return lines[at]
}

func paperWidthOf(columns int) int64 {
	if columns == 32 {
		return 58
	}

	return 80
}
