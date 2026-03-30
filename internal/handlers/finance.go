package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/finance"
)

const maxGlobalSymbols = 20

// FinanceHandler handles finance/stock-related HTTP requests.
type FinanceHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

// GetSymbolCount returns the total number of finance symbols across all widgets.
func (h *FinanceHandler) GetSymbolCount(w http.ResponseWriter, r *http.Request) {
	count := h.countGlobalSymbols(r.Context())
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

// GetQuotes returns real-time quotes for the requested symbols.
func (h *FinanceHandler) GetQuotes(w http.ResponseWriter, r *http.Request) {
	symbolsParam := r.URL.Query().Get("symbols")
	if symbolsParam == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "missing symbols parameter"})
		return
	}

	// Parse and normalize symbols
	raw := strings.Split(symbolsParam, ",")
	symbols := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(strings.ToUpper(s))
		if s != "" {
			symbols = append(symbols, s)
		}
	}
	if len(symbols) == 0 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "no valid symbols"})
		return
	}
	if len(symbols) > maxGlobalSymbols {
		symbols = symbols[:maxGlobalSymbols]
	}

	// Build deterministic cache key
	sorted := make([]string, len(symbols))
	copy(sorted, symbols)
	sort.Strings(sorted)
	cacheKey := "finance:quotes:" + strings.Join(sorted, ",")

	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client := finance.NewClient(nil)
	data, err := client.GetQuotes(r.Context(), symbols)
	if err != nil {
		slog.Error("finance: failed to fetch quotes", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch quotes"})
		return
	}

	h.Cache.SetWithTTL(cacheKey, data, 60*time.Second)
	writeJSON(w, http.StatusOK, data)
}

// SearchSymbol searches for symbols matching a query (autocomplete).
func (h *FinanceHandler) SearchSymbol(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "missing q parameter"})
		return
	}

	cacheKey := "finance:search:" + strings.ToLower(query)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client := finance.NewClient(nil)
	results, err := client.Search(r.Context(), query)
	if err != nil {
		slog.Error("finance: failed to search symbols", "query", query, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to search"})
		return
	}

	h.Cache.SetWithTTL(cacheKey, results, 5*time.Minute)
	writeJSON(w, http.StatusOK, results)
}

// countGlobalSymbols counts unique symbols across all finance widgets.
func (h *FinanceHandler) countGlobalSymbols(ctx context.Context) int {
	rows, err := h.DB.QueryContext(ctx,
		"SELECT config FROM widgets WHERE type = 'finance'",
	)
	if err != nil {
		return 0
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	for rows.Next() {
		var configStr string
		if err := rows.Scan(&configStr); err != nil {
			continue
		}
		var cfg struct {
			Symbols []string `json:"symbols"`
		}
		if err := json.Unmarshal([]byte(configStr), &cfg); err != nil {
			continue
		}
		for _, s := range cfg.Symbols {
			seen[strings.ToUpper(s)] = struct{}{}
		}
	}
	return len(seen)
}
