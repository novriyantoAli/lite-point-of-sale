// Package struk holds the Struk domain: the printable receipt of one Penjualan,
// as lines of text wrapped to the paper width (CONTEXT.md, Struk).
//
// It is the one place that decides *what the Struk says* — which lines, in which
// order, wrapped how — and nothing about how those lines reach a printer. The
// ESC/POS adapter translates the lines to bytes (ADR-0017, keputusan 2), so the
// domain can be tested without a device and the adapter can be tested against
// lines that are known.
package struk

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
)

// Template is the part of the store's Pengaturan a Struk is printed from: the
// header and footer blocks an Admin edits, and the paper width they are wrapped
// to (ADR-0017, keputusan 4).
//
// It is deliberately not domainpengaturan.Settings: a Struk cares about the
// template alone, and taking the settings row would drag the ambang Stok menipis
// — a setting with nothing to do with printing — into this package.
type Template struct {
	// Header and Footer are free-text blocks, newline-separated, printed verbatim
	// above and below the sale lines (CONTEXT.md, Struk).
	Header string
	Footer string
	// PaperWidth is the thermal roll width in millimetres — 58 or 80.
	PaperWidth int64
}

// Columns is the number of printable columns of a paper width: 58 mm → 32,
// 80 mm → 48.
//
// The column count is derived from the width here, never stored as a second
// value that could drift (ADR-0017, keputusan 4). The stored Pengaturan only
// ever holds 58 or 80, so anything else falls back to the wider roll rather than
// wrapping a receipt to a column count nobody chose.
func Columns(paperWidth int64) int {
	if paperWidth == 58 {
		return 32
	}

	return 48
}

// Compose renders the Struk of one Penjualan against the store's template.
//
// The content follows the order CONTEXT.md lists (ADR-0017, keputusan 4): the
// header block, the Nomor Struk, when it was rung up, the Kasir, the Items, the
// total, the Pembayaran method, the amount paid and — for Tunai only — the
// Kembalian, then the footer block.
//
// Long lines are wrapped to the paper width, never truncated: a Produk name that
// does not fit loses information if it is cut, and a Struk that is wrong is worse
// than a Struk that is taller.
func Compose(sale domainpenjualan.Sale, template Template) []string {
	columns := Columns(template.PaperWidth)

	var lines []string

	// The header block is free text, so a blank line separates it from the sale.
	if block := blockLines(template.Header, columns); len(block) > 0 {
		lines = append(lines, block...)
		lines = append(lines, "")
	}

	lines = append(lines,
		fmt.Sprintf("Nomor Struk: %d", sale.ReceiptNumber),
		sale.CreatedAt,
		fmt.Sprintf("Kasir: %s", sale.CashierName),
	)

	for _, item := range sale.Items {
		lines = append(lines, wrapText(item.Name, columns)...)
		lines = append(lines, moneyLine(
			fmt.Sprintf("  %d x %s", item.Quantity, formatMoney(item.Price)),
			formatMoney(item.Subtotal()),
			columns,
		)...)
	}

	lines = append(lines, moneyLine("Total", formatMoney(sale.Total), columns)...)
	lines = append(lines, moneyLine(
		fmt.Sprintf("Bayar (%s)", methodLabel(sale.Payment.Method)),
		formatMoney(sale.Payment.Amount),
		columns,
	)...)

	// Only Tunai has a Kembalian: a recorded method pays the total exactly, so
	// printing a Kembalian of zero would be a line that means nothing
	// (CONTEXT.md, Kembalian).
	if sale.Payment.Method == domainpenjualan.PaymentCash {
		lines = append(lines, moneyLine("Kembalian", formatMoney(sale.Payment.Change), columns)...)
	}

	if block := blockLines(template.Footer, columns); len(block) > 0 {
		lines = append(lines, "")
		lines = append(lines, block...)
	}

	return lines
}

// blockLines renders one free-text template block: each of its own lines,
// wrapped to the paper width. An empty block is no lines at all, so Compose can
// tell "the Admin left it blank" from "the Admin typed a blank line".
func blockLines(block string, columns int) []string {
	block = strings.TrimRight(block, "\n")
	if strings.TrimSpace(block) == "" {
		return nil
	}

	var lines []string
	for _, line := range strings.Split(block, "\n") {
		lines = append(lines, wrapText(line, columns)...)
	}

	return lines
}

// wrapText breaks one line of text into lines of at most `columns` characters,
// at a space where it can and mid-word where it must.
//
// A line that already fits is answered untouched — including its own spacing —
// so a template block the Admin laid out is printed as typed. Only a line that
// is too long is reflowed, and a long first word is broken rather than allowed
// to run off the roll.
func wrapText(text string, columns int) []string {
	if columns <= 0 || utf8.RuneCountInString(text) <= columns {
		return []string{text}
	}

	// A block the Admin indented stays indented on every wrapped piece.
	indent := text[:len(text)-len(strings.TrimLeft(text, " "))]
	if utf8.RuneCountInString(indent) >= columns {
		indent = ""
	}
	// What is left of the roll once the indent is taken out.
	width := columns - utf8.RuneCountInString(indent)

	var (
		lines   []string
		current string
	)
	flush := func() {
		if current != "" {
			lines = append(lines, current)
			current = ""
		}
	}

	for _, word := range strings.Fields(text) {
		// A single word longer than what is left of the roll is cut at the width:
		// there is nowhere else for it to go, and a line wider than the paper
		// prints clipped anyway.
		for utf8.RuneCountInString(word) > width {
			flush()
			runes := []rune(word)
			lines = append(lines, indent+string(runes[:width]))
			word = string(runes[width:])
		}

		if current == "" {
			current = indent + word
			continue
		}
		if utf8.RuneCountInString(current)+1+utf8.RuneCountInString(word) <= columns {
			current += " " + word
			continue
		}

		flush()
		current = indent + word
	}
	flush()

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

// moneyLine puts `left` on the left and `right` on the right, padded to the
// paper width. If the two do not fit on one line, `left` is wrapped and `right`
// gets a line of its own, right-aligned: the money is never dropped, only moved
// (ADR-0017, keputusan 4).
func moneyLine(left, right string, columns int) []string {
	leftLen := utf8.RuneCountInString(left)
	rightLen := utf8.RuneCountInString(right)

	if leftLen+1+rightLen <= columns {
		return []string{left + strings.Repeat(" ", columns-leftLen-rightLen) + right}
	}

	lines := wrapText(left, columns)

	padding := columns - rightLen
	if padding < 0 {
		padding = 0
	}

	return append(lines, strings.Repeat(" ", padding)+right)
}

// methodLabel is how a Pembayaran method is written for the person holding the
// Struk. The Go identifiers are English and the DTOs stay English (ADR-0012);
// what is *printed* is the Indonesian domain term.
func methodLabel(method domainpenjualan.PaymentMethod) string {
	switch method {
	case domainpenjualan.PaymentCash:
		return "Tunai"
	case domainpenjualan.PaymentQRIS:
		return "QRIS"
	case domainpenjualan.PaymentDebit:
		return "Debit"
	case domainpenjualan.PaymentTransfer:
		return "Transfer"
	default:
		return string(method)
	}
}

// formatMoney renders whole rupiah the way the app writes them for a person:
// thousands grouped with a dot and an "Rp " prefix. Money has no decimals in
// this app (CONTEXT.md), so there is no fraction to render.
func formatMoney(amount int64) string {
	return "Rp " + groupThousands(amount)
}

func groupThousands(amount int64) string {
	digits := strconv.FormatInt(amount, 10)

	sign := ""
	if strings.HasPrefix(digits, "-") {
		sign, digits = "-", digits[1:]
	}

	var grouped strings.Builder
	for i, digit := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			grouped.WriteByte('.')
		}
		grouped.WriteRune(digit)
	}

	return sign + grouped.String()
}
