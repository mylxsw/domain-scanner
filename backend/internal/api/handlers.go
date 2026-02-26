package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mylxsw/namecheap-domain-probe/backend/internal/model"
	"github.com/mylxsw/namecheap-domain-probe/backend/internal/service"
)

// Handler holds all API handlers
type Handler struct {
	probeService *service.ProbeService
	tldService   *service.TLDService
	cfg          *model.NamecheapConfig
}

// NewHandler creates a new handler
func NewHandler(probeService *service.ProbeService, tldService *service.TLDService, cfg *model.NamecheapConfig) *Handler {
	return &Handler{
		probeService: probeService,
		tldService:   tldService,
		cfg:          cfg,
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	})
}

// CreateProbe handles POST /api/probe
func (h *Handler) CreateProbe(c *gin.Context) {
	var req model.CreateProbeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	task, err := h.probeService.CreateTask(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "create_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// ListProbes handles GET /api/probe
func (h *Handler) ListProbes(c *gin.Context) {
	limit := 50
	offset := 0
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	items, err := h.probeService.ListTasks(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "list_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"limit":  limit,
		"offset": offset,
		"count":  len(items),
	})
}

// GetProbeStatus handles GET /api/probe/:id
func (h *Handler) GetProbeStatus(c *gin.Context) {
	taskID := c.Param("id")
	task, ok := h.probeService.GetTask(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetProbeResults handles GET /api/probe/:id/results
func (h *Handler) GetProbeResults(c *gin.Context) {
	taskID := c.Param("id")
	task, ok := h.probeService.GetTask(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "Task not found",
		})
		return
	}

	results, _, err := h.probeService.GetResultsOrLoad(taskID)
	if err != nil {
		log.Printf("load probe results failed for task=%s: %v", taskID, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"task":    task,
		"results": results,
	})
}

// StreamProbe handles GET /api/probe/:id/stream
func (h *Handler) StreamProbe(c *gin.Context) {
	taskID := c.Param("id")
	task, ok := h.probeService.GetTask(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "Task not found",
		})
		return
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// Create a channel for progress updates
	progressChan := make(chan interface{}, 100)

	// Start the task if it's pending
	if task.Status == model.ProbeStatusPending {
		go func() {
			defer close(progressChan)
			// Use background context instead of request context
			ctx := context.Background()
			_ = h.probeService.RunTask(ctx, taskID, progressChan)
		}()

		// Stream progress events
		for obj := range progressChan {
			data := toJSON(obj)

			sse := model.SSEEvent{Event: "message", Data: data}
			fmt.Fprint(c.Writer, sse.String())
			c.Writer.Flush()
		}
	} else if task.Status == model.ProbeStatusRunning {
		// Task is already running, poll for progress
		// Send initial status
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		lastCompleted := 0
		for {
			select {
			case <-c.Request.Context().Done():
				return
			case <-ticker.C:
				currentTask, _ := h.probeService.GetTask(taskID)
				if currentTask == nil {
					return
				}

				// Send progress update if changed
				if currentTask.Completed > lastCompleted {
					lastCompleted = currentTask.Completed
					progress := model.ProgressEvent{
						Type:  "progress",
						Index: currentTask.Completed,
						Total: currentTask.Total,
					}
					sse := model.SSEEvent{Event: "message", Data: toJSON(progress)}
					fmt.Fprint(c.Writer, sse.String())
					c.Writer.Flush()
				}

				// Check if task completed or failed
				if currentTask.Status == model.ProbeStatusCompleted || currentTask.Status == model.ProbeStatusFailed {
					// Send final results wrapped in ProgressEvent
					results, _ := h.probeService.GetResults(taskID)
					for i, result := range results {
						progressEvent := model.ProgressEvent{
							Type:  "progress",
							Index: i + 1,
							Total: len(results),
							Data:  result,
						}
						event := model.SSEEvent{
							Event: "message",
							Data:  toJSON(progressEvent),
						}
						fmt.Fprint(c.Writer, event.String())
						c.Writer.Flush()
					}

					// Send summary
					summary := model.SummaryEvent{
						Type:           "summary",
						Word:           currentTask.Word,
						Total:          currentTask.Total,
						Outdir:         currentTask.Outdir,
						ReportMd:       currentTask.ReportMd,
						ResultsCsv:     currentTask.ResultsCsv,
						ResultsJsonl:   currentTask.ResultsJsonl,
						FinishedAt:     currentTask.FinishedAt.Format("2006-01-02 15:04:05"),
						ElapsedSeconds: currentTask.ElapsedSeconds,
					}
					event := model.SSEEvent{
						Event: "message",
						Data:  toJSON(summary),
					}
					fmt.Fprint(c.Writer, event.String())
					c.Writer.Flush()
					return
				}
			}
		}
	} else {
		// Task is completed or failed, send cached results wrapped in ProgressEvent
		results, _ := h.probeService.GetResults(taskID)
		for i, result := range results {
			progressEvent := model.ProgressEvent{
				Type:  "progress",
				Index: i + 1,
				Total: len(results),
				Data:  result,
			}
			event := model.SSEEvent{
				Event: "message",
				Data:  toJSON(progressEvent),
			}
			fmt.Fprint(c.Writer, event.String())
			c.Writer.Flush()
		}

		// Send summary
		summary := model.SummaryEvent{
			Type:           "summary",
			Word:           task.Word,
			Total:          task.Total,
			Outdir:         task.Outdir,
			ReportMd:       task.ReportMd,
			ResultsCsv:     task.ResultsCsv,
			ResultsJsonl:   task.ResultsJsonl,
			FinishedAt:     task.FinishedAt.Format("2006-01-02 15:04:05"),
			ElapsedSeconds: task.ElapsedSeconds,
		}
		event := model.SSEEvent{
			Event: "message",
			Data:  toJSON(summary),
		}
		fmt.Fprint(c.Writer, event.String())
		c.Writer.Flush()
		return
	}
}

// GetTLDs handles GET /api/tlds
func (h *Handler) GetTLDs(c *gin.Context) {
	tldMode := c.Query("mode")
	mode := model.TldModeAll
	switch tldMode {
	case "api-registerable-only":
		mode = model.TldModeApiRegisterableOnly
	case "mainstream-only":
		mode = model.TldModeMainstreamOnly
	}

	tlds, err := h.tldService.GetTLDsByMode(c.Request.Context(), mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tlds":  tlds,
		"count": len(tlds),
	})
}

// toJSON converts an object to JSON string
func toJSON(obj interface{}) string {
	data, _ := json.Marshal(obj)
	return string(data)
}
