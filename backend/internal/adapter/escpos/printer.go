package escpos

import (
	"context"
	"fmt"
	"os"
	"strings"

	domainstruk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/struk"
)

// New answers the printer the store has configured.
//
// An empty device means no printer is set up: the answer is a null printer that
// reports "printer belum diatur" as a print failure, rather than a panic or a
// success nobody can tell apart from a real print (ADR-0017, keputusan 2). A
// device that is set is opened per print, so a printer unplugged between sales is
// a failure the next sale reports instead of a stale file handle.
func New(device string) domainstruk.Printer {
	if strings.TrimSpace(device) == "" {
		return NullPrinter{}
	}

	return DevicePrinter{Device: device}
}

// DevicePrinter writes the bytes of a Struk to a device path, such as
// /dev/usb/lp0. It satisfies domain/struk.Printer, so no ESC/POS byte leaves this
// package (ADR-0004).
type DevicePrinter struct {
	// Device is the path passed to POS_PRINTER_DEVICE.
	Device string
}

// Print opens the device, writes one Struk and closes it. A device that cannot
// be opened or written is answered with an error naming the path — the Kasir
// reads a message, not a stack trace.
func (p DevicePrinter) Print(ctx context.Context, lines []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	device, err := os.OpenFile(p.Device, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return fmt.Errorf("open printer %s: %w", p.Device, err)
	}
	defer device.Close()

	if _, err := device.Write(Encode(lines)); err != nil {
		return fmt.Errorf("write to printer %s: %w", p.Device, err)
	}

	return nil
}

// NullPrinter is the printer of a store that has not configured one. Every print
// fails with domainstruk.ErrPrinterNotConfigured, which is the honest answer:
// there is no device, so no Struk came out.
type NullPrinter struct{}

// Print always fails with ErrPrinterNotConfigured.
func (NullPrinter) Print(context.Context, []string) error {
	return domainstruk.ErrPrinterNotConfigured
}
