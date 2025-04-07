package request

import (
	"covalence/src/types"
	"time"
)

// RequestMetrics collects metrics about the request
type Metrics struct {
	StartTime              time.Time
	RequestPreparationTime time.Duration
	HookTime               time.Duration
	RequestBodyTime        time.Duration
	UpstreamLatency        time.Duration
	TotalProcessTime       time.Duration
	StatusCode             int
	Name                   types.Name
	Model                  types.ModelID
	StreamingResponse      bool
}
