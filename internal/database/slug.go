package database

import (
	"database/sql"
	"fmt"
	"strings"
	"unicode"
)

// translitMap maps common European diacritics to ASCII equivalents.
var translitMap = map[rune]string{
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a",
	'æ': "ae",
	'ç': "c",
	'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
	'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
	'ð': "d",
	'ñ': "n",
	'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o",
	'ù': "u", 'ú': "u", 'û': "u", 'ü': "u",
	'ý': "y", 'ÿ': "y",
	'ß': "ss",
	'œ': "oe",
	'À': "a", 'Á': "a", 'Â': "a", 'Ã': "a", 'Ä': "a", 'Å': "a",
	'Æ': "ae",
	'Ç': "c",
	'È': "e", 'É': "e", 'Ê': "e", 'Ë': "e",
	'Ì': "i", 'Í': "i", 'Î': "i", 'Ï': "i",
	'Ð': "d",
	'Ñ': "n",
	'Ò': "o", 'Ó': "o", 'Ô': "o", 'Õ': "o", 'Ö': "o", 'Ø': "o",
	'Ù': "u", 'Ú': "u", 'Û': "u", 'Ü': "u",
	'Ý': "y",
	'Œ': "oe",
}

const maxSlugLength = 100

// GenerateSlug converts a name into a URL-friendly slug.
func GenerateSlug(name string) string {
	name = strings.ToLower(name)

	var b strings.Builder
	for _, r := range name {
		if repl, ok := translitMap[r]; ok {
			b.WriteString(repl)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}

	slug := b.String()

	// Collapse consecutive dashes
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")

	if slug == "" {
		slug = "category"
	}

	// Truncate at maxSlugLength, preferring a dash boundary
	if len(slug) > maxSlugLength {
		slug = slug[:maxSlugLength]
		if idx := strings.LastIndex(slug, "-"); idx > maxSlugLength/2 {
			slug = slug[:idx]
		}
	}

	return slug
}

// generateUniqueSlug returns a slug that doesn't exist in the categories table.
func (db *DB) generateUniqueSlug(base string) (string, error) {
	slug := base
	for i := 2; i <= 99; i++ {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM categories WHERE slug = ?", slug).Scan(&count); err != nil {
			return "", fmt.Errorf("checking slug uniqueness: %w", err)
		}
		if count == 0 {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	return slug, nil
}

// generateUniqueSlugTx returns a unique slug within a transaction context.
func generateUniqueSlugTx(tx *sql.Tx, base string) (string, error) {
	slug := base
	for i := 2; i <= 99; i++ {
		var count int
		if err := tx.QueryRow("SELECT COUNT(*) FROM categories WHERE slug = ?", slug).Scan(&count); err != nil {
			return "", fmt.Errorf("checking slug uniqueness: %w", err)
		}
		if count == 0 {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
	return slug, nil
}
