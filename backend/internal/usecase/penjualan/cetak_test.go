package penjualan

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	domainpengaturan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/pengaturan"
	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
	domainstruk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/struk"
)

// nomorStrukLine renders the Struk line that names one Nomor Struk.
func nomorStrukLine(receiptNumber int64) string {
	return "Nomor Struk: " + strconv.FormatInt(receiptNumber, 10)
}

func containsLine(lines []string, want string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, want) {
			return true
		}
	}

	return false
}

func TestCheckoutPrintsTheStrukOfTheSaleItStored(t *testing.T) {
	products := newFakeProducts(kopi(), teh())
	sales := newFakeSales()
	settings := newFakeSettings()
	printer := &fakePrinter{}

	result, err := NewService(products, sales, settings, printer).Checkout(context.Background(), kasir, CheckoutInput{
		Items: []ItemInput{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 3},
		},
		Payment: PaymentInput{Method: "cash", Amount: 60000},
	})
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}

	if !result.Print.Printed {
		t.Fatalf("print result: got %+v, want a printed Struk", result.Print)
	}
	if result.Print.Message != "" {
		t.Errorf("print message: got %q, want none on success", result.Print.Message)
	}
	if len(printer.printed) != 1 {
		t.Fatalf("printed: got %d Struk, want 1", len(printer.printed))
	}

	struk := printer.printed[0]

	// The Struk names the sale the checkout just stored — the Nomor Struk the
	// repository assigned, not one the use case invented.
	if !containsLine(struk, nomorStrukLine(result.Sale.ReceiptNumber)) {
		t.Errorf("Struk %q: want the stored Nomor Struk", struk)
	}
	if !containsLine(struk, "Kasir: kasir1") {
		t.Errorf("Struk %q: want the Kasir's name", struk)
	}
	if !containsLine(struk, "Kopi Susu") || !containsLine(struk, "Teh Botol") {
		t.Errorf("Struk %q: want both Items", struk)
	}
	if !containsLine(struk, "Total") || !containsLine(struk, "Kembalian") {
		t.Errorf("Struk %q: want the total and the Tunai Kembalian", struk)
	}
}

func TestCheckoutPrintsWithTheStoredTemplate(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()
	settings := newFakeSettings()
	settings.settings = domainpengaturan.Settings{
		ID:         1,
		Header:     "Toko Kopi Purnama\nJl. Melati 1",
		Footer:     "Terima kasih",
		PaperWidth: 58,
	}
	printer := &fakePrinter{}

	if _, err := NewService(products, sales, settings, printer).Checkout(
		context.Background(), kasir, tunai(1, 1, 18000),
	); err != nil {
		t.Fatalf("checkout: %v", err)
	}

	struk := printer.printed[0]

	if !containsLine(struk, "Toko Kopi Purnama") || !containsLine(struk, "Jl. Melati 1") {
		t.Errorf("Struk %q: want the header block", struk)
	}
	if !containsLine(struk, "Terima kasih") {
		t.Errorf("Struk %q: want the footer block", struk)
	}
	// The stored 58 mm is what wraps the Struk: 32 columns, not the wider roll.
	for _, line := range struk {
		if len([]rune(line)) > 32 {
			t.Errorf("Struk at 58 mm: line %q is wider than 32 columns", line)
		}
	}
}

// TestCheckoutKeepsTheSaleWhenThePrinterFails is the decision of ADR-0017,
// keputusan 1: the Penjualan is money that has already moved, so a printer that
// does not answer must never turn a recorded sale into a failed checkout.
func TestCheckoutKeepsTheSaleWhenThePrinterFails(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()
	printer := &fakePrinter{err: errors.New("perangkat tidak ditemukan")}

	result, err := NewService(products, sales, newFakeSettings(), printer).Checkout(
		context.Background(), kasir, tunai(1, 1, 18000),
	)
	if err != nil {
		t.Fatalf("checkout with a broken printer: got %v, want the sale to survive", err)
	}

	// The sale is stored, and the answer says the Struk did not come out.
	if len(sales.recorded) != 1 {
		t.Fatalf("recorded: got %d Penjualan, want 1", len(sales.recorded))
	}
	if result.Sale.ReceiptNumber != 1 {
		t.Errorf("Nomor Struk: got %d, want the stored 1", result.Sale.ReceiptNumber)
	}
	if result.Print.Printed {
		t.Error("print result: got printed, want a reported failure")
	}
	if result.Print.Message == "" {
		t.Error("print message: got nothing, want a reason the Kasir can read")
	}
	if len(printer.printed) != 0 {
		t.Errorf("printed: got %d Struk, want none", len(printer.printed))
	}
}

func TestCheckoutSaysWhenNoPrinterIsConfigured(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()
	printer := &fakePrinter{err: domainstruk.ErrPrinterNotConfigured}

	result, err := NewService(products, sales, newFakeSettings(), printer).Checkout(
		context.Background(), kasir, tunai(1, 1, 18000),
	)
	if err != nil {
		t.Fatalf("checkout without a printer: got %v, want the sale to survive", err)
	}

	if result.Print.Printed {
		t.Error("print result: got printed, want a reported failure")
	}
	if result.Print.Message != "Printer belum diatur." {
		t.Errorf("print message: got %q, want the message a Kasir can act on", result.Print.Message)
	}
}

func TestCheckoutKeepsTheSaleWhenTheTemplateCannotBeRead(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()
	settings := newFakeSettings()
	settings.err = errors.New("database sedang sibuk")

	result, err := NewService(products, sales, settings, &fakePrinter{}).Checkout(
		context.Background(), kasir, tunai(1, 1, 18000),
	)
	if err != nil {
		t.Fatalf("checkout with unreadable Pengaturan: got %v, want the sale to survive", err)
	}

	if len(sales.recorded) != 1 {
		t.Fatalf("recorded: got %d Penjualan, want 1", len(sales.recorded))
	}
	if result.Print.Printed {
		t.Error("print result: got printed, want a reported failure")
	}
	// The message does not blame the printer: nothing was sent to one.
	if result.Print.Message != composeFailureMessage {
		t.Errorf("print message: got %q, want the message for a Struk that could not be composed", result.Print.Message)
	}
}

func TestCetakStrukPrintsAStoredSale(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()
	printer := &fakePrinter{}
	service := NewService(products, sales, newFakeSettings(), printer)

	created := saleOf(t, service, tunai(1, 1, 18000))

	printed, err := service.PrintReceipt(context.Background(), created.ReceiptNumber)
	if err != nil {
		t.Fatalf("PrintReceipt: %v", err)
	}

	if !printed.Printed {
		t.Fatalf("print result: got %+v, want a printed Struk", printed)
	}
	if len(printer.printed) != 2 {
		t.Fatalf("printed: got %d Struk, want the checkout's and the reprint's", len(printer.printed))
	}

	reprint := printer.printed[1]
	if !containsLine(reprint, nomorStrukLine(created.ReceiptNumber)) {
		t.Errorf("reprint %q: want the same Nomor Struk", reprint)
	}
}

// TestCetakStrukUsesTheCurrentTemplate is keputusan 5 of ADR-0017: a reprint
// prints the store's template as it stands now, not a copy from when the sale
// happened.
func TestCetakStrukUsesTheCurrentTemplate(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()
	settings := newFakeSettings()
	printer := &fakePrinter{}
	service := NewService(products, sales, settings, printer)

	created := saleOf(t, service, tunai(1, 1, 18000))

	settings.settings.Header = "Toko Kopi Purnama"

	if _, err := service.PrintReceipt(context.Background(), created.ReceiptNumber); err != nil {
		t.Fatalf("PrintReceipt: %v", err)
	}

	reprint := printer.printed[1]
	if !containsLine(reprint, "Toko Kopi Purnama") {
		t.Errorf("reprint %q: want the template as it stands now", reprint)
	}
	// The original print, made before the rename, did not have it.
	if containsLine(printer.printed[0], "Toko Kopi Purnama") {
		t.Errorf("first print %q: want the template of its own moment", printer.printed[0])
	}
}

func TestCetakStrukReportsAMissingSale(t *testing.T) {
	service := NewService(newFakeProducts(), newFakeSales(), newFakeSettings(), &fakePrinter{})

	_, err := service.PrintReceipt(context.Background(), 404)
	if !errors.Is(err, domainpenjualan.ErrSaleNotFound) {
		t.Fatalf("got %v, want ErrSaleNotFound", err)
	}
}

func TestCetakStrukReportsAPrinterFailureWithoutAnError(t *testing.T) {
	products := newFakeProducts(kopi())
	sales := newFakeSales()
	printer := &fakePrinter{err: domainstruk.ErrPrinterNotConfigured}
	service := NewService(products, sales, newFakeSettings(), printer)

	created := saleOf(t, service, tunai(1, 1, 18000))

	printed, err := service.PrintReceipt(context.Background(), created.ReceiptNumber)
	if err != nil {
		t.Fatalf("PrintReceipt with a broken printer: got %v, want a reported failure", err)
	}
	if printed.Printed || printed.Message != "Printer belum diatur." {
		t.Errorf("print result: got %+v, want the unconfigured-printer message", printed)
	}
}
