package escpos

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	domainstruk "github.com/novriyantoAli/lite-point-of-sale/backend/internal/domain/struk"
)

func TestNewWithoutADeviceAnswersTheNullPrinter(t *testing.T) {
	for _, device := range []string{"", "   "} {
		if _, ok := New(device).(NullPrinter); !ok {
			t.Errorf("New(%q): got %T, want a NullPrinter", device, New(device))
		}
	}
}

func TestNullPrinterReportsAFailureRatherThanSucceedingSilently(t *testing.T) {
	err := NullPrinter{}.Print(context.Background(), []string{"Nomor Struk: 1"})

	if !errors.Is(err, domainstruk.ErrPrinterNotConfigured) {
		t.Fatalf("Print: got %v, want ErrPrinterNotConfigured", err)
	}
}

func TestDevicePrinterWritesTheEncodedStrukToThePath(t *testing.T) {
	// A file in t.TempDir() stands in for the device: the bytes that would reach
	// the printer are read back and asserted (ADR-0007). The file is created first
	// because a device path is never created by the adapter — a typo must fail
	// loudly rather than silently print to a new file.
	device := filepath.Join(t.TempDir(), "printer.bin")
	if err := os.WriteFile(device, nil, 0o600); err != nil {
		t.Fatalf("create the device file: %v", err)
	}

	lines := []string{"Nomor Struk: 12", "Total   Rp 54.000"}
	if err := New(device).Print(context.Background(), lines); err != nil {
		t.Fatalf("Print: %v", err)
	}

	printed, err := os.ReadFile(device)
	if err != nil {
		t.Fatalf("read the device file: %v", err)
	}
	if !bytes.Equal(printed, Encode(lines)) {
		t.Errorf("printed bytes: got % X, want % X", printed, Encode(lines))
	}
}

func TestDevicePrinterAppendsEachStruk(t *testing.T) {
	device := filepath.Join(t.TempDir(), "printer.bin")
	if err := os.WriteFile(device, nil, 0o600); err != nil {
		t.Fatalf("create the device file: %v", err)
	}

	printer := New(device)
	if err := printer.Print(context.Background(), []string{"Nomor Struk: 1"}); err != nil {
		t.Fatalf("first Print: %v", err)
	}
	if err := printer.Print(context.Background(), []string{"Nomor Struk: 2"}); err != nil {
		t.Fatalf("second Print: %v", err)
	}

	printed, err := os.ReadFile(device)
	if err != nil {
		t.Fatalf("read the device file: %v", err)
	}
	// Two Struk came out; the second did not overwrite the first.
	if !bytes.Contains(printed, []byte("Nomor Struk: 1\n")) || !bytes.Contains(printed, []byte("Nomor Struk: 2\n")) {
		t.Errorf("printed bytes % X: want both Struk", printed)
	}
}

func TestDevicePrinterReportsADeviceItCannotOpen(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "tidak-ada.bin")

	err := New(missing).Print(context.Background(), []string{"Nomor Struk: 1"})

	if err == nil {
		t.Fatal("Print: got no error, want a failure naming the device")
	}
	if !bytes.Contains([]byte(err.Error()), []byte(missing)) {
		t.Errorf("Print: got %q, want it to name %q", err, missing)
	}
}

func TestDevicePrinterRespectsACancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	device := filepath.Join(t.TempDir(), "printer.bin")
	if err := os.WriteFile(device, nil, 0o600); err != nil {
		t.Fatalf("create the device file: %v", err)
	}

	if err := New(device).Print(ctx, []string{"Nomor Struk: 1"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("Print with a cancelled context: got %v, want context.Canceled", err)
	}
}
