package database

import "testing"

func TestGenerateSlug(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple ascii", "My Category", "my-category"},
		{"french accents", "Mon Salon", "mon-salon"},
		{"accented chars", "Réseaux et Sécurité", "reseaux-et-securite"},
		{"cedilla", "Français", "francais"},
		{"german", "Straße", "strasse"},
		{"ligature oe", "Cœur", "coeur"},
		{"multiple spaces", "my   category", "my-category"},
		{"special chars", "A!@#$%^&*()B", "a-b"},
		{"leading trailing dashes", "--test--", "test"},
		{"empty string", "", "category"},
		{"only symbols", "!@#$%", "category"},
		{"numbers", "Room 42", "room-42"},
		{"mixed", "Café & Réseau #1", "cafe-reseau-1"},
		{"uppercase accents", "ÉLÉPHANT", "elephant"},
		{"ae ligature", "Ærodynamic", "aerodynamic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GenerateSlug(tt.input)
			if got != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGenerateSlugTruncation(t *testing.T) {
	t.Parallel()
	long := ""
	for i := 0; i < 30; i++ {
		long += "abcde-"
	}
	slug := GenerateSlug(long)
	if len(slug) > 100 {
		t.Errorf("slug length %d exceeds max 100", len(slug))
	}
}
