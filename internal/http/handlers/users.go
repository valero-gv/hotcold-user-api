package handlers

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"azret/internal/model"
	"azret/internal/service"
)

var userIDRe = regexp.MustCompile(`^[A-Za-z0-9_\-]{1,20}$`)

type MetaData struct {
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

type ColdCacheResponse struct {
	Code         int      `json:"code"`
	Message      string   `json:"message"`
	UserID       *string  `json:"user_id,omitempty"`
	Deeplink     *string  `json:"deeplink,omitempty"`
	PromoMessage *string  `json:"promo_message,omitempty"`
	CacheHit     bool     `json:"cache_hit"`
	Metadata     MetaData `json:"metadata"`
}

type UsersHandler struct {
	svc    *service.UserService
	logger *zap.Logger
}

func NewUsersHandler(svc *service.UserService, logger *zap.Logger) *UsersHandler {
	return &UsersHandler{svc: svc, logger: logger}
}

// GetUser handles GET /v1/users?user_id=...
// - On cache hit (hot storage: Redis) returns { message: "ok:redis", cache_hit: true }
// - On cache miss (cold storage: PostgreSQL) returns { message: "ok:postgres", cache_hit: false }
// - Validation errors -> 400, Not found -> 404, Internal -> 500
func (h *UsersHandler) GetUser(c *gin.Context) {
	start := time.Now()
	id := c.Query("user_id")
	if !userIDRe.MatchString(id) {
		c.JSON(http.StatusBadRequest, ColdCacheResponse{
			Code:     http.StatusBadRequest,
			Message:  "invalid user_id",
			CacheHit: false,
			Metadata: MetaData{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "v1"},
		})
		return
	}

	u, cacheHit, found, err := h.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("get_user_error", zap.String("user_id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, ColdCacheResponse{
			Code:     http.StatusInternalServerError,
			Message:  "internal error",
			CacheHit: cacheHit,
			Metadata: MetaData{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "v1"},
		})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, ColdCacheResponse{
			Code:     http.StatusNotFound,
			Message:  "not found",
			CacheHit: cacheHit,
			Metadata: MetaData{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "v1"},
		})
		return
	}

	// Success: message reflects the data source; code mirrors HTTP 200
	msg := "ok:postgres"
	if cacheHit {
		msg = "ok:redis"
	}
	resp := toResponse(u, cacheHit, http.StatusOK, msg)
	c.JSON(http.StatusOK, resp)
	h.logger.Info("get_user",
		zap.String("user_id", id),
		zap.Bool("cache_hit", cacheHit),
		zap.Int("code", resp.Code),
		zap.Duration("latency_ms", time.Since(start)),
	)
}

func toResponse(u model.User, cacheHit bool, code int, message string) ColdCacheResponse {
	ts := time.Now().UTC().Format(time.RFC3339)
	return ColdCacheResponse{
		Code:         code,
		Message:      message,
		UserID:       strPtr(u.UserID),
		Deeplink:     strPtr(u.Deeplink),
		PromoMessage: strPtr(u.PromoMessage),
		CacheHit:     cacheHit,
		Metadata:     MetaData{Timestamp: ts, Version: "v1"},
	}
}

func strPtr(s string) *string { return &s }
