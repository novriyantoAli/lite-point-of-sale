package e2e

import (
	"bytes"
	"net/http"
	"os"
	"strings"
	"testing"
)

// printPayload is the outcome of printing one Struk, as the API answers it on
// both paths: the automatic print a checkout reports, and the reprint endpoint.
type printPayload struct {
	Printed bool   `json:"printed"`
	Message string `json:"message"`
}

// checkoutResultEnvelope is the checkout answer: the stored Penjualan plus the
// result of the Struk print that followed it (ADR-0017, keputusan 1).
type checkoutResultEnvelope struct {
	Sale  salePayload  `json:"sale"`
	Print printPayload `json:"print"`
}

// printEnvelope is the reprint answer.
type printEnvelope struct {
	Print printPayload `json:"print"`
}

// newPrinterStore boots the real service with a printer file the test can read
// back, and logs in as the seeded Admin. The file is what stands in for the
// thermal printer: the bytes that would have come out of it are read here
// (ADR-0017, keputusan 2).
func newPrinterStore(t *testing.T) (baseURL, token, printerPath string) {
	t.Helper()

	cfg := newTestConfig(t)
	_, baseURL = startAPI(t, cfg)

	return baseURL, logIn(t, baseURL, seededAdmin, testAdminPassword), cfg.PrinterDevice
}

// checkoutPrinting posts a cart and answers the stored Penjualan together with
// the print result the checkout reported.
func checkoutPrinting(t *testing.T, baseURL, token string, payload checkoutPayload) (salePayload, printPayload) {
	t.Helper()

	var created dataEnvelope[checkoutResultEnvelope]
	status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan", token, payload, &created)
	if status != http.StatusCreated {
		t.Fatalf("checkout: got status %d, want %d", status, http.StatusCreated)
	}

	return created.Data.Sale, created.Data.Print
}

// reprint asks the API to print the Struk of one stored Penjualan and answers
// the print result, failing the test when the API refused it.
func reprint(t *testing.T, baseURL, token string, receiptNumber int64) printPayload {
	t.Helper()

	var printed dataEnvelope[printEnvelope]
	status := apiCall(t, http.MethodPost, baseURL+"/api/penjualan/"+itoa(receiptNumber)+"/struk", token, nil, &printed)
	if status != http.StatusOK {
		t.Fatalf("reprint Penjualan %d: got status %d, want %d", receiptNumber, status, http.StatusOK)
	}

	return printed.Data.Print
}

// printerSince reads the bytes the printer file received after `from`, and
// answers them with the file's new length. It is how a test isolates the Struk of
// one print from the ones before it.
func printerSince(t *testing.T, path string, from int) (string, int) {
	t.Helper()

	printed, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read the printer file: %v", err)
	}
	if from > len(printed) {
		from = 0
	}

	return string(printed[from:]), len(printed)
}

// assertEscposWrapped checks the bytes really are one ESC/POS job: the printer is
// initialised at the start and the paper is cut at the end.
func assertEscposWrapped(t *testing.T, printed string) {
	t.Helper()

	if !strings.HasPrefix(printed, "\x1b@") {
		t.Errorf("printed bytes %q: want them to start with the initialise command", printed)
	}
	if !strings.HasSuffix(printed, "\x1dVB\x00") {
		t.Errorf("printed bytes %q: want them to end with the cut command", printed)
	}
}

// TestCheckoutPrintsTheStrukAndReportsIt is the automatic print of ADR-0017: a
// checkout stores the Penjualan and prints its Struk, and the answer carries the
// outcome of that print.
func TestCheckoutPrintsTheStrukAndReportsIt(t *testing.T) {
	baseURL, token, printerPath := newPrinterStore(t)

	// A header and footer, so the printed Struk shows the template as well as the
	// sale.
	var settings dataEnvelope[pengaturanEnvelope]
	if status := apiCall(t, http.MethodPut, baseURL+"/api/pengaturan", token, pengaturanPayload{
		Header:            "Toko Kopi Purnama",
		Footer:            "Terima kasih",
		PaperWidth:        80,
		LowStockThreshold: 5,
	}, &settings); status != http.StatusOK {
		t.Fatalf("change Pengaturan: got status %d, want %d", status, http.StatusOK)
	}

	kopi := createProduk(t, baseURL, token, produkPayload{Name: "Kopi Susu", Price: 18000, Stock: 10})
	teh := createProduk(t, baseURL, token, produkPayload{Name: "Teh Botol", Price: 6000, Stock: 5})

	sale, printed := checkoutPrinting(t, baseURL, token, checkoutPayload{
		Items: []checkoutItemPayload{
			{ProductID: kopi.ID, Quantity: 2},
			{ProductID: teh.ID, Quantity: 1},
		},
		Payment: checkoutPaymentPayload{Method: "cash", Amount: 50000},
	})

	if !printed.Printed {
		t.Fatalf("print result: got %+v, want the Struk to have printed", printed)
	}
	if printed.Message != "" {
		t.Errorf("print message: got %q, want none on success", printed.Message)
	}

	raw, _ := printerSince(t, printerPath, 0)
	assertEscposWrapped(t, raw)

	// The content CONTEXT.md lists: the template blocks, the Nomor Struk, who rang
	// it up, the Items at the price they were sold for, the total, the method and
	// the Tunai Kembalian (ADR-0017, keputusan 4).
	for _, want := range []string{
		"Toko Kopi Purnama",
		"Terima kasih",
		"Nomor Struk: " + itoa(sale.ReceiptNumber),
		"Kasir: " + seededAdmin,
		"Kopi Susu",
		"  2 x Rp 18.000",
		"Rp 36.000",
		"Teh Botol",
		"  1 x Rp 6.000",
		"Rp 6.000",
		"Total",
		"Rp 42.000",
		"Bayar (Tunai)",
		"Rp 50.000",
		"Kembalian",
		"Rp 8.000",
	} {
		if !strings.Contains(raw, want) {
			t.Errorf("printed Struk %q: want it to contain %q", raw, want)
		}
	}
}

// TestCheckoutKeepsTheSaleWhenNoPrinterIsConfigured is keputusan 1 and 2 of
// ADR-0017: with no POS_PRINTER_DEVICE there is no printer, the print fails with
// a message a Kasir can read, and the Penjualan is stored anyway.
func TestCheckoutKeepsTheSaleWhenNoPrinterIsConfigured(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.PrinterDevice = ""
	_, baseURL := startAPI(t, cfg)
	token := logIn(t, baseURL, seededAdmin, testAdminPassword)

	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})

	sale, printed := checkoutPrinting(t, baseURL, token, tunai(created.ID, 1, 18000))

	if printed.Printed {
		t.Error("print result: got printed, want a reported failure without a printer")
	}
	if printed.Message != "Printer belum diatur." {
		t.Errorf("print message: got %q, want the message a Kasir can act on", printed.Message)
	}

	// The sale survived: it is readable by its Nomor Struk, and the Stok left.
	stored := readSale(t, baseURL, token, sale.ReceiptNumber)
	if stored.Total != 18000 {
		t.Errorf("stored sale: got %+v, want the Penjualan the checkout recorded", stored)
	}
	if stock := stockOf(t, baseURL, token, created.ID); stock != 4 {
		t.Errorf("Stok after a checkout whose print failed: got %d, want 4", stock)
	}
}

// TestReprintPrintsTheStrukOfAStoredSale is the reprint path of ADR-0017,
// keputusan 5: the same use case the checkout runs, reached by Nomor Struk.
func TestReprintPrintsTheStrukOfAStoredSale(t *testing.T) {
	baseURL, token, printerPath := newPrinterStore(t)

	var settings dataEnvelope[pengaturanEnvelope]
	if status := apiCall(t, http.MethodPut, baseURL+"/api/pengaturan", token, pengaturanPayload{
		Header:            "Toko Kopi Purnama",
		Footer:            "Terima kasih",
		PaperWidth:        58,
		LowStockThreshold: 5,
	}, &settings); status != http.StatusOK {
		t.Fatalf("change Pengaturan: got status %d, want %d", status, http.StatusOK)
	}

	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi Susu", Price: 18000, Stock: 5})
	sale, _ := checkoutPrinting(t, baseURL, token, tunai(created.ID, 2, 40000))

	// Only the reprint's own bytes: the checkout already printed one Struk.
	_, offset := printerSince(t, printerPath, 0)

	printed := reprint(t, baseURL, token, sale.ReceiptNumber)

	if !printed.Printed {
		t.Fatalf("reprint: got %+v, want the Struk to have printed", printed)
	}

	raw, _ := printerSince(t, printerPath, offset)
	assertEscposWrapped(t, raw)

	for _, want := range []string{
		"Nomor Struk: " + itoa(sale.ReceiptNumber),
		"Kopi Susu",
		"  2 x Rp 18.000",
		"Rp 36.000",
		"Total",
		"Bayar (Tunai)",
		"Kembalian",
		"Rp 4.000",
		"Toko Kopi Purnama",
		"Terima kasih",
	} {
		if !strings.Contains(raw, want) {
			t.Errorf("reprinted Struk %q: want it to contain %q", raw, want)
		}
	}

	// At 58 mm the Struk is wrapped to 32 columns, not the 48 of the wider roll.
	for _, line := range strings.Split(raw, "\n") {
		if len([]rune(line)) > 32 {
			t.Errorf("reprinted Struk at 58 mm: line %q is wider than 32 columns", line)
		}
	}
}

// TestARecordedMethodReprintsWithoutAKembalian checks the one rule that differs
// between the methods on a printed Struk (CONTEXT.md, Kembalian).
func TestARecordedMethodReprintsWithoutAKembalian(t *testing.T) {
	baseURL, token, printerPath := newPrinterStore(t)

	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi QRIS", Price: 7000, Stock: 5})
	sale, printed := checkoutPrinting(t, baseURL, token, checkoutPayload{
		Items:   []checkoutItemPayload{{ProductID: created.ID, Quantity: 1}},
		Payment: checkoutPaymentPayload{Method: "qris", Amount: 7000},
	})
	if !printed.Printed {
		t.Fatalf("checkout print: got %+v, want it to have printed", printed)
	}

	_, offset := printerSince(t, printerPath, 0)
	reprint(t, baseURL, token, sale.ReceiptNumber)
	raw, _ := printerSince(t, printerPath, offset)

	if !strings.Contains(raw, "Bayar (QRIS)") {
		t.Errorf("printed Struk %q: want the method named", raw)
	}
	if strings.Contains(raw, "Kembalian") {
		t.Errorf("printed Struk %q: want no Kembalian for a recorded method", raw)
	}
}

// TestReprintUsesTheCurrentTemplate is keputusan 5 of ADR-0017: the template
// belongs to the store, so a reprint prints it as it stands now rather than a
// copy from when the sale happened.
func TestReprintUsesTheCurrentTemplate(t *testing.T) {
	baseURL, token, printerPath := newPrinterStore(t)

	var settings dataEnvelope[pengaturanEnvelope]
	put := func(header string) {
		t.Helper()
		if status := apiCall(t, http.MethodPut, baseURL+"/api/pengaturan", token, pengaturanPayload{
			Header:            header,
			Footer:            "",
			PaperWidth:        80,
			LowStockThreshold: 5,
		}, &settings); status != http.StatusOK {
			t.Fatalf("change Pengaturan: got status %d, want %d", status, http.StatusOK)
		}
	}

	put("Toko Lama")
	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})
	sale, _ := checkoutPrinting(t, baseURL, token, tunai(created.ID, 1, 18000))

	// The shop renames itself after the sale.
	put("Toko Baru")

	_, offset := printerSince(t, printerPath, 0)
	reprint(t, baseURL, token, sale.ReceiptNumber)
	raw, _ := printerSince(t, printerPath, offset)

	if !strings.Contains(raw, "Toko Baru") {
		t.Errorf("reprinted Struk %q: want the template as it stands now", raw)
	}
	if strings.Contains(raw, "Toko Lama") {
		t.Errorf("reprinted Struk %q: want no copy of the template from when the sale happened", raw)
	}
}

func TestReprintNeedsASessionButNotAnAdmin(t *testing.T) {
	baseURL, adminToken, _ := newPrinterStore(t)

	created := createProduk(t, baseURL, adminToken, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})
	sale, _ := checkoutPrinting(t, baseURL, adminToken, tunai(created.ID, 1, 18000))

	t.Run("anonymously", func(t *testing.T) {
		var failure errorPayload
		status := apiCall(t, http.MethodPost,
			baseURL+"/api/penjualan/"+itoa(sale.ReceiptNumber)+"/struk", "", nil, &failure)

		if status != http.StatusUnauthorized {
			t.Fatalf("got status %d, want %d", status, http.StatusUnauthorized)
		}
		if failure.Error != "invalid_token" {
			t.Errorf("got code %q, want %q", failure.Error, "invalid_token")
		}
	})

	t.Run("as Kasir", func(t *testing.T) {
		// CONTEXT.md gives the Kasir "cetak Struk", so this route carries no role
		// guard — the same as the checkout and the sale read.
		createPengguna(t, baseURL, adminToken, "kasir1", kasirPassword, "kasir")
		kasirToken := logIn(t, baseURL, "kasir1", kasirPassword)

		if printed := reprint(t, baseURL, kasirToken, sale.ReceiptNumber); !printed.Printed {
			t.Errorf("reprint as Kasir: got %+v, want it to have printed", printed)
		}
	})
}

func TestReprintReportsAMissingOrUnusableNomorStruk(t *testing.T) {
	baseURL, token, _ := newPrinterStore(t)

	tests := []struct {
		name          string
		path          string
		wantStatus    int
		wantErrorCode string
	}{
		{
			name: "a Penjualan that is not there", path: "/api/penjualan/404/struk",
			wantStatus: http.StatusNotFound, wantErrorCode: "sale_not_found",
		},
		{
			name: "a Nomor Struk that is not a number", path: "/api/penjualan/abc/struk",
			wantStatus: http.StatusBadRequest, wantErrorCode: "invalid_input",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var failure errorPayload
			status := apiCall(t, http.MethodPost, baseURL+test.path, token, nil, &failure)

			if status != test.wantStatus {
				t.Fatalf("got status %d, want %d", status, test.wantStatus)
			}
			if failure.Error != test.wantErrorCode {
				t.Errorf("got code %q, want %q", failure.Error, test.wantErrorCode)
			}
		})
	}
}

// TestPrintedStrukIsRawBytesNoClientCouldRead pins the transport: the bytes on
// the wire are ESC/POS commands, not text a browser would have produced
// (ADR-0003).
func TestPrintedStrukIsRawBytesNoClientCouldRead(t *testing.T) {
	baseURL, token, printerPath := newPrinterStore(t)

	created := createProduk(t, baseURL, token, produkPayload{Name: "Kopi", Price: 18000, Stock: 5})
	checkoutPrinting(t, baseURL, token, tunai(created.ID, 1, 18000))

	printed, err := os.ReadFile(printerPath)
	if err != nil {
		t.Fatalf("read the printer file: %v", err)
	}

	if !bytes.Contains(printed, []byte{0x1B, 0x40}) {
		t.Errorf("printed bytes % X: want the ESC/POS initialise command", printed)
	}
	if !bytes.Contains(printed, []byte{0x1B, 0x45, 0x00}) {
		t.Errorf("printed bytes % X: want the emphasis-off command", printed)
	}
}
