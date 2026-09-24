package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	domainauth "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/auth"
	domainpenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/penjualan"
	usecasepenjualan "github.com/novriyantoAli/lite-point-of-sale/backend/internal/usecase/penjualan"
)

// SaleService is the Penjualan use cases the HTTP adapter depends on. Declaring
// it here, next to the handlers that use it, keeps this package testable with a
// fake service and keeps the dependency pointing inward (ADR-0004).
type SaleService interface {
	Checkout(ctx context.Context, cashier domainauth.PublicUser, input usecasepenjualan.CheckoutInput) (usecasepenjualan.CheckoutResult, error)
	FindByReceiptNumber(ctx context.Context, receiptNumber int64) (domainpenjualan.Sale, error)
	PrintReceipt(ctx context.Context, receiptNumber int64) (usecasepenjualan.PrintResult, error)
	ListSales(ctx context.Context, filter domainpenjualan.ReportFilter) ([]domainpenjualan.SaleSummary, error)
	DailyRevenue(ctx context.Context, filter domainpenjualan.ReportFilter) (domainpenjualan.DailyRevenue, error)
}

// checkoutRequest is the cart a Kasir posts. The Items carry the Produk id and
// the quantity only: the name and the price are the catalogue's to know, and a
// client that sent them would be able to write its own history.
type checkoutRequest struct {
	Items   []checkoutItemRequest  `json:"items"`
	Payment checkoutPaymentRequest `json:"payment"`
}

type checkoutItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int64 `json:"quantity"`
}

// checkoutPaymentRequest is the Pembayaran being recorded. Amount is what the
// buyer handed over for Tunai; the Kembalian is the API's to work out, never the
// client's to declare.
type checkoutPaymentRequest struct {
	Method string `json:"method"`
	Amount int64  `json:"amount"`
}

func (r checkoutRequest) input() usecasepenjualan.CheckoutInput {
	items := make([]usecasepenjualan.ItemInput, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, usecasepenjualan.ItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return usecasepenjualan.CheckoutInput{
		Items: items,
		Payment: usecasepenjualan.PaymentInput{
			Method: r.Payment.Method,
			Amount: r.Payment.Amount,
		},
	}
}

// saleResponse is the JSON view of a Penjualan: the Nomor Struk, who rang it up,
// what was bought at the price it was bought for, and how it was paid.
type saleResponse struct {
	ReceiptNumber int64              `json:"receipt_number"`
	CreatedAt     string             `json:"created_at"`
	CashierID     int64              `json:"cashier_id"`
	CashierName   string             `json:"cashier_name"`
	Total         int64              `json:"total"`
	Items         []saleItemResponse `json:"items"`
	Payment       paymentResponse    `json:"payment"`
}

// saleItemResponse is one Item as it was recorded: the Produk it came from, plus
// the name and price copied at checkout.
type saleItemResponse struct {
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Subtotal  int64  `json:"subtotal"`
}

// paymentResponse is the Pembayaran of a Penjualan. For Tunai, `change` is the
// Kembalian; for a non-tunai method it is zero.
type paymentResponse struct {
	Method string `json:"method"`
	Amount int64  `json:"amount"`
	Change int64  `json:"change"`
}

// saleEnvelope wraps one stored Penjualan: the answer to reading a sale by its
// Nomor Struk.
type saleEnvelope struct {
	Sale saleResponse `json:"sale"`
}

// printResponse is the outcome of printing one Struk. `printed` is what the till
// branches on; `message` is the reason a Kasir can read when it is false. It is
// the same shape on both print paths — the automatic print a checkout answers and
// the reprint endpoint (ADR-0017, keputusan 5).
type printResponse struct {
	Printed bool   `json:"printed"`
	Message string `json:"message,omitempty"`
}

// checkoutEnvelope wraps the Penjualan a checkout stored together with the
// outcome of the Struk print that followed it. The print is reported, never
// allowed to fail the sale: the money has already moved (ADR-0017, keputusan 1).
type checkoutEnvelope struct {
	Sale  saleResponse  `json:"sale"`
	Print printResponse `json:"print"`
}

// printEnvelope wraps the outcome of a reprint, which is all the `/penjualan`
// screen needs: it already has the sale on screen.
type printEnvelope struct {
	Print printResponse `json:"print"`
}

// saleSummaryResponse is one row of the sales list (#9): the Nomor Struk and the
// few fields the list shows, without the Item lines a detail read carries.
type saleSummaryResponse struct {
	ReceiptNumber int64  `json:"receipt_number"`
	CreatedAt     string `json:"created_at"`
	CashierID     int64  `json:"cashier_id"`
	CashierName   string `json:"cashier_name"`
	Total         int64  `json:"total"`
	Method        string `json:"method"`
}

// methodTotalResponse is what one Pembayaran method contributed to a day: how
// many Penjualan and how much.
type methodTotalResponse struct {
	Method       string `json:"method"`
	Total        int64  `json:"total"`
	Transactions int64  `json:"transactions"`
}

// cashierTotalResponse is what one Kasir rang up in a day.
type cashierTotalResponse struct {
	CashierID    int64  `json:"cashier_id"`
	CashierName  string `json:"cashier_name"`
	Total        int64  `json:"total"`
	Transactions int64  `json:"transactions"`
}

// revenueResponse is the omzet of one store-local day (#9): the total, the number
// of Penjualan, and the breakdown by Pembayaran method and by Kasir.
type revenueResponse struct {
	Date         string                 `json:"date"`
	Total        int64                  `json:"total"`
	Transactions int64                  `json:"transactions"`
	ByMethod     []methodTotalResponse  `json:"by_method"`
	ByCashier    []cashierTotalResponse `json:"by_cashier"`
}

func newPrintResponse(print usecasepenjualan.PrintResult) printResponse {
	return printResponse{Printed: print.Printed, Message: print.Message}
}

func newSaleResponse(sale domainpenjualan.Sale) saleResponse {
	items := make([]saleItemResponse, 0, len(sale.Items))
	for _, item := range sale.Items {
		items = append(items, saleItemResponse{
			ProductID: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
			Subtotal:  item.Subtotal(),
		})
	}

	return saleResponse{
		ReceiptNumber: sale.ReceiptNumber,
		CreatedAt:     sale.CreatedAt,
		CashierID:     sale.CashierID,
		CashierName:   sale.CashierName,
		Total:         sale.Total,
		Items:         items,
		Payment: paymentResponse{
			Method: string(sale.Payment.Method),
			Amount: sale.Payment.Amount,
			Change: sale.Payment.Change,
		},
	}
}

// newSaleSummaryResponses renders the sales list of a day. It answers an empty
// list rather than null for a day with no sales, so the screen can iterate it
// without a guard.
func newSaleSummaryResponses(sales []domainpenjualan.SaleSummary) []saleSummaryResponse {
	responses := make([]saleSummaryResponse, 0, len(sales))
	for _, sale := range sales {
		responses = append(responses, saleSummaryResponse{
			ReceiptNumber: sale.ReceiptNumber,
			CreatedAt:     sale.CreatedAt,
			CashierID:     sale.CashierID,
			CashierName:   sale.CashierName,
			Total:         sale.Total,
			Method:        string(sale.Method),
		})
	}

	return responses
}

// newRevenueResponse renders the omzet of a day, with both breakdowns.
func newRevenueResponse(report domainpenjualan.DailyRevenue) revenueResponse {
	methods := make([]methodTotalResponse, 0, len(report.ByMethod))
	for _, total := range report.ByMethod {
		methods = append(methods, methodTotalResponse{
			Method:       string(total.Method),
			Total:        total.Total,
			Transactions: total.Transactions,
		})
	}

	cashiers := make([]cashierTotalResponse, 0, len(report.ByCashier))
	for _, total := range report.ByCashier {
		cashiers = append(cashiers, cashierTotalResponse{
			CashierID:    total.CashierID,
			CashierName:  total.CashierName,
			Total:        total.Total,
			Transactions: total.Transactions,
		})
	}

	return revenueResponse{
		Date:         report.Date,
		Total:        report.Total,
		Transactions: report.Transactions,
		ByMethod:     methods,
		ByCashier:    cashiers,
	}
}

// checkoutHandler turns a cart into a Penjualan: the atomic write that records
// the sale, snapshots each Item's name and price, and takes the Stok out
// (CONTEXT.md, Penjualan).
//
// The answer carries the sale and the outcome of the Struk print that followed
// it. A print that failed is reported here, not turned into an error: the
// Penjualan is stored either way (ADR-0017, keputusan 1).
//
// Any signed-in Pengguna may ring one up. Both Peran of CONTEXT.md sell at a
// one-terminal store — the Admin is the owner behind the counter as often as the
// Kasir is — so this route is the one part of the API without a role guard.
func checkoutHandler(service SaleService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cashier, ok := requireUser(w, r, logger)
		if !ok {
			return
		}

		var request checkoutRequest
		if !decodeJSON(w, r, &request, logger) {
			return
		}

		result, err := service.Checkout(r.Context(), cashier, request.input())
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusCreated, dataResponse{Data: checkoutEnvelope{
			Sale:  newSaleResponse(result.Sale),
			Print: newPrintResponse(result.Print),
		}}, logger)
	}
}

// saleHandler answers one stored Penjualan by its Nomor Struk. The Nomor Struk
// is what the Struk prints, so it is what a reprint or a look-up arrives with.
func saleHandler(service SaleService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		receiptNumber, ok := receiptNumberParam(w, r, logger)
		if !ok {
			return
		}

		sale, err := service.FindByReceiptNumber(r.Context(), receiptNumber)
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: saleEnvelope{
			Sale: newSaleResponse(sale),
		}}, logger)
	}
}

// printSaleHandler prints the Struk of one stored Penjualan: the reprint the till
// offers after a print failed, and the one `/penjualan` offers for an old sale.
//
// It is the same use case the checkout runs automatically, so the answer is the
// same shape as the print result a checkout reports (ADR-0017, keputusan 5). Like
// the checkout and the sale read, it carries no role guard: CONTEXT.md gives the
// Kasir "cetak Struk".
func printSaleHandler(service SaleService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		receiptNumber, ok := receiptNumberParam(w, r, logger)
		if !ok {
			return
		}

		printed, err := service.PrintReceipt(r.Context(), receiptNumber)
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: printEnvelope{
			Print: newPrintResponse(printed),
		}}, logger)
	}
}

// receiptNumberParam reads the {receiptNumber} of the route. A path that is not
// a number at all is invalid input, not a missing Penjualan.
func receiptNumberParam(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (int64, bool) {
	receiptNumber, err := strconv.ParseInt(r.PathValue("receiptNumber"), 10, 64)
	if err != nil || receiptNumber <= 0 {
		writeInvalidInput(w, "Nomor Struk tidak valid.", logger)
		return 0, false
	}

	return receiptNumber, true
}

// listSalesHandler answers the Penjualan of one store-local day: the sales list
// an Admin reads, newest Nomor Struk first. The day is `?date=YYYY-MM-DD` and
// defaults to today, which the use case resolves (ADR-0015).
//
// It is a literal next to the `{receiptNumber}` wildcard of the sale read, and a
// literal is the more specific pattern — the same reason `/api/produk/kategori`
// wins over `/api/produk/{id}`. "omzet" is not a Nomor Struk either.
func listSalesHandler(service SaleService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sales, err := service.ListSales(r.Context(), reportFilter(r))
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: newSaleSummaryResponses(sales)}, logger)
	}
}

// dailyRevenueHandler answers the omzet of one store-local day: the total, the
// number of Penjualan, and the breakdown by Pembayaran method and by Kasir. Like
// the sales list it is Admin-only and defaults to today.
func dailyRevenueHandler(service SaleService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report, err := service.DailyRevenue(r.Context(), reportFilter(r))
		if err != nil {
			writeError(w, err, logger)
			return
		}

		writeJSON(w, http.StatusOK, dataResponse{Data: newRevenueResponse(report)}, logger)
	}
}

// reportFilter reads the day of a report off the query string. An absent date is
// left empty for the use case to read as today; a malformed one is refused there
// with a message, rather than silently matching nothing here.
func reportFilter(r *http.Request) domainpenjualan.ReportFilter {
	return domainpenjualan.ReportFilter{Date: strings.TrimSpace(r.URL.Query().Get("date"))}
}
