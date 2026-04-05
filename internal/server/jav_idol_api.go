package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common"
	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/manager"
)

func listJavIdols(c *gin.Context) {
	limit := queryInt(c, "limit", 100)
	offset := queryInt(c, "offset", 0)
	search := strings.TrimSpace(c.Query("search"))
	sort := strings.TrimSpace(c.Query("sort"))

	items, total, err := dbpkg.ListJavIdols(c.Request.Context(), search, sort, limit, offset)
	if err != nil {
		logging.Error("list jav idols: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	cfg := common.AppConfig
	coverDir := ""
	if cfg != nil {
		coverDir = cfg.JavCoverDir
	}
	for i := range items {
		items[i].Name = strings.TrimSpace(items[i].Name)
		items[i].RomanName = strings.TrimSpace(items[i].RomanName)
		items[i].JapaneseName = strings.TrimSpace(items[i].JapaneseName)
		items[i].ChineseName = strings.TrimSpace(items[i].ChineseName)

		// If current sample cover not present, try another code preferring solo works.
		if coverDir == "" {
			continue
		}
		if _, ok := manager.FindCoverPath(coverDir, items[i].SampleCode); ok {
			continue
		}
		codes, err := dbpkg.ListIdolCoverCodes(c.Request.Context(), items[i].ID)
		if err != nil {
			logging.Error("list idol cover codes id=%d: %v", items[i].ID, err)
			continue
		}
		var chosen string
		for _, code := range codes {
			if _, ok := manager.FindCoverPath(coverDir, code); ok {
				chosen = code
				break
			}
		}
		if chosen == "" && len(codes) > 0 {
			chosen = codes[0]
		}
		if chosen != "" {
			items[i].SampleCode = chosen
			if common.CoverManager != nil && !common.CoverManager.Exists(chosen) {
				common.CoverManager.Enqueue(chosen)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
	})
}
