package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/superindexer/indexer/internal/database/models"
	"github.com/superindexer/indexer/internal/http/validator"
	"github.com/uptrace/bun"
)

// SettingsHandler handles system settings endpoints.
type SettingsHandler struct {
	DB   *bun.DB
	Log  interface{}
}

// UpdateSettingRequest represents a request to update a setting.
type UpdateSettingRequest struct {
	Value interface{} `json:"value" validate:"required"`
}

// List handles listing all settings.
func (h *SettingsHandler) List(c echo.Context) error {
	var settings []models.SystemSetting
	err := h.DB.NewSelect().
		Model(&settings).
		Order("setting_key ASC").
		Scan(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list settings: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": settings,
	})
}

// Get handles getting a specific setting.
func (h *SettingsHandler) Get(c echo.Context) error {
	key := c.Param("key")

	var setting models.SystemSetting
	err := h.DB.NewSelect().
		Model(&setting).
		Where("setting_key = ?", key).
		Scan(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "setting not found")
	}

	var value interface{}
	if err := json.Unmarshal([]byte(setting.SettingValue), &value); err != nil {
		value = setting.SettingValue
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"setting": map[string]interface{}{
			"key":   setting.SettingKey,
			"value": value,
		},
	})
}

// Update handles updating a specific setting.
func (h *SettingsHandler) Update(c echo.Context) error {
	key := c.Param("key")

	var req UpdateSettingRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	// Marshal the value to JSON
	valueJSON, err := json.Marshal(req.Value)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid setting value: "+err.Error())
	}

	// Check if setting exists
	var existing models.SystemSetting
	err = h.DB.NewSelect().
		Model(&existing).
		Where("setting_key = ?", key).
		Scan(c.Request().Context())
	if err != nil {
		// Setting doesn't exist, create it
		setting := models.SystemSetting{
			SettingKey:   key,
			SettingValue: string(valueJSON),
		}
		_, err = h.DB.NewInsert().Model(&setting).Exec(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create setting: "+err.Error())
		}
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"setting": map[string]interface{}{
				"key":   key,
				"value": req.Value,
			},
		})
	}

	// Update existing setting
	_, err = h.DB.NewUpdate().
		Model(&models.SystemSetting{}).
		Set("setting_value = ?", string(valueJSON)).
		Where("setting_key = ?", key).
		Exec(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update setting: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"setting": map[string]interface{}{
			"key":   key,
			"value": req.Value,
		},
	})
}