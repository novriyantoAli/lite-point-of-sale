package pengaturan

import "testing"

// StoreName is the rule the login screen depends on before a session exists:
// the store's name is the first non-empty line of the Struk header block. These
// cases pin the edges — leading blank lines, whitespace-only lines, and an
// empty block.
func TestSettingsStoreName(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "first line is the name",
			header: "Toko Elektrik Maju\nJl. Merdeka No. 12\nTelp 0812-3456-7890",
			want:   "Toko Elektrik Maju",
		},
		{
			name:   "blank lines before the name are skipped",
			header: "\n\n  Toko Elektrik Maju  \nJl. Merdeka No. 12",
			want:   "Toko Elektrik Maju",
		},
		{
			name:   "an empty block has no name",
			header: "",
			want:   "",
		},
		{
			name:   "whitespace-only lines still have no name",
			header: "   \n\t\n  ",
			want:   "",
		},
		{
			name:   "the name is trimmed but never reflowed",
			header: "Toko Elektrik Maju",
			want:   "Toko Elektrik Maju",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			settings := Settings{Header: test.header}
			if got := settings.StoreName(); got != test.want {
				t.Errorf("StoreName() = %q, want %q", got, test.want)
			}
		})
	}
}
