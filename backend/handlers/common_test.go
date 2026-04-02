package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"prompt-vault/middleware"
	"prompt-vault/models"
)

func TestParsePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		queryParams map[string]string
		wantPage    int
		wantLimit   int
	}{
		{
			name:        "default values",
			queryParams: map[string]string{},
			wantPage:    1,
			wantLimit:   20,
		},
		{
			name:        "custom page and limit",
			queryParams: map[string]string{"page": "3", "limit": "10"},
			wantPage:    3,
			wantLimit:   10,
		},
		{
			name:        "page less than 1 defaults to 1",
			queryParams: map[string]string{"page": "0"},
			wantPage:    1,
			wantLimit:   20,
		},
		{
			name:        "negative page defaults to 1",
			queryParams: map[string]string{"page": "-5"},
			wantPage:    1,
			wantLimit:   20,
		},
		{
			name:        "limit less than 1 defaults to 20",
			queryParams: map[string]string{"limit": "0"},
			wantPage:    1,
			wantLimit:   20,
		},
		{
			name:        "limit greater than max defaults to 100",
			queryParams: map[string]string{"limit": "500"},
			wantPage:    1,
			wantLimit:   100,
		},
		{
			name:        "limit at max boundary",
			queryParams: map[string]string{"limit": "100"},
			wantPage:    1,
			wantLimit:   100,
		},
		{
			name:        "non-numeric page defaults",
			queryParams: map[string]string{"page": "abc"},
			wantPage:    1,
			wantLimit:   20,
		},
		{
			name:        "non-numeric limit defaults",
			queryParams: map[string]string{"limit": "xyz"},
			wantPage:    1,
			wantLimit:   20,
		},
		{
			name:        "first page calculation",
			queryParams: map[string]string{"page": "1", "limit": "10"},
			wantPage:    1,
			wantLimit:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			router := gin.New()
			var gotOffset, gotPage, gotLimit int

			router.GET("/test", func(c *gin.Context) {
				countQ := db.Model(&models.Prompt{})
				orderedQ := db.Model(&models.Prompt{})
				offset, page, limit, _, _ := middleware.ParsePagination(c, countQ, orderedQ)
				gotOffset, gotPage, gotLimit = offset, page, limit
				c.JSON(http.StatusOK, gin.H{
					"page":   page,
					"limit":  limit,
					"offset": offset,
				})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for k, v := range tt.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			if gotPage != tt.wantPage {
				t.Errorf("page: got %d, want %d", gotPage, tt.wantPage)
			}
			if gotLimit != tt.wantLimit {
				t.Errorf("limit: got %d, want %d", gotLimit, tt.wantLimit)
			}

			// Verify offset calculation: (page-1) * limit
			wantOffset := (tt.wantPage - 1) * tt.wantLimit
			if gotOffset != wantOffset {
				t.Errorf("offset: got %d, want %d", gotOffset, wantOffset)
			}
		})
	}
}

func TestParsePagination_OffsetCalculation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		page   int
		limit  int
		offset int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 20, 40},
		{1, 10, 0},
		{5, 10, 40},
		{1, 100, 0},
		{10, 50, 450},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			db := newTestDB(t)
			router := gin.New()
			var gotOffset int

			router.GET("/test", func(c *gin.Context) {
				countQ := db.Model(&models.Prompt{})
				orderedQ := db.Model(&models.Prompt{})
				offset, _, _, _, _ := middleware.ParsePagination(c, countQ, orderedQ)
				gotOffset = offset
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.URL.RawQuery = "page=" + itoa(c.page) + "&limit=" + itoa(c.limit)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if gotOffset != c.offset {
				t.Errorf("page=%d limit=%d: offset got %d, want %d", c.page, c.limit, gotOffset, c.offset)
			}
		})
	}
}

func TestParsePagination_PaginationMeta(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		page      int
		limit     int
		wantPages int
	}{
		{1, 10, 0},
		{1, 25, 0},
		{1, 5, 0},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			db := newTestDB(t)
			router := gin.New()
			var gotMeta models.PaginationMeta

			router.GET("/test", func(c *gin.Context) {
				countQ := db.Model(&models.Prompt{})
				orderedQ := db.Model(&models.Prompt{})
				_, _, _, _, meta := middleware.ParsePagination(c, countQ, orderedQ)
				gotMeta = meta
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.URL.RawQuery = "page=" + itoa(c.page) + "&limit=" + itoa(c.limit)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if gotMeta.TotalPages != c.wantPages {
				t.Errorf("page=%d limit=%d: totalPages got %d, want %d", c.page, c.limit, gotMeta.TotalPages, c.wantPages)
			}
			if gotMeta.Page != c.page {
				t.Errorf("page: got %d, want %d", gotMeta.Page, c.page)
			}
			if gotMeta.Limit != c.limit {
				t.Errorf("limit: got %d, want %d", gotMeta.Limit, c.limit)
			}
		})
	}
}

func TestCalcTotalPages(t *testing.T) {
	cases := []struct {
		total int64
		limit int
		pages int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{25, 10, 3},
		{100, 10, 10},
		{101, 10, 11},
		{1, 0, 0},
		{1, -5, 0},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			got := middleware.CalcTotalPages(c.total, c.limit)
			if got != c.pages {
				t.Errorf("total=%d limit=%d: got %d, want %d", c.total, c.limit, got, c.pages)
			}
		})
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}
