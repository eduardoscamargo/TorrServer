package api

import (
	"net/http"

	"server/bonjour"
	"server/dlna"
	"server/rutor"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	sets "server/settings"
	"server/torr"
)

// Action: get, set, def, speed_get, speed_set, speed_restore
type setsReqJS struct {
	requestI
	Sets              *sets.BTSets `json:"sets,omitempty"`
	DownloadRateLimit *int         `json:"download_rate_limit,omitempty"`
	Temporary         bool         `json:"temporary,omitempty"`
}

// settings godoc
//
//	@Summary		Get / Set server settings
//	@Description	Get or set server settings. speed_get returns configured/effective download limits in KB/s; speed_set changes the live limit using download_rate_limit and optional temporary; speed_restore removes a temporary override.
//
//	@Tags			API
//
//	@Param			request	body	setsReqJS	true	"Settings request. Available actions: get, set, def, speed_get, speed_set, speed_restore"
//
//	@Accept			json
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	sets.BTSets	"Settings JSON or nothing. Depends on what action has been asked."
//	@Router			/settings [post]
func settings(c *gin.Context) {
	var req setsReqJS
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Action == "get" {
		c.JSON(200, sets.BTsets)
		return
	} else if req.Action == "speed_get" {
		c.JSON(http.StatusOK, torr.GetDownloadSpeedLimit())
		return
	} else if req.Action == "speed_set" {
		if req.DownloadRateLimit == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "download_rate_limit is required"})
			return
		}
		state, err := torr.UpdateDownloadSpeedLimit(*req.DownloadRateLimit, req.Temporary)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, state)
		return
	} else if req.Action == "speed_restore" {
		state, err := torr.RestoreDownloadSpeedLimit()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, state)
		return
	} else if req.Action == "set" {
		if req.Sets == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "sets is required"})
			return
		}
		torr.SetSettings(req.Sets)
		dlna.Stop()
		if req.Sets.EnableDLNA {
			dlna.Start()
		}
		bonjour.Stop()
		if req.Sets.EnableBonjour {
			bonjour.Start()
		}
		rutor.Stop()
		rutor.Start()
		c.Status(200)
		return
	} else if req.Action == "def" {
		torr.SetDefSettings()
		dlna.Stop()
		bonjour.Stop()
		rutor.Stop()
		c.Status(200)
		return
	}
	c.AbortWithError(http.StatusBadRequest, errors.New("action is empty"))
}
