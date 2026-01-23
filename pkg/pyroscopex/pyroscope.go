package pyroscopex

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/grafana/pyroscope-go"
)

// profiler holds the singleton instance of pyroscope profiler
var profiler *pyroscope.Profiler

// once ensures that the profiler is initialized only once
var once sync.Once

// Start initializes and starts the pyroscope continuous profiling
// It sets up mutex and block profiling rates if configured
// The profiler will only be started once regardless of how many times this function is called
func Start(config *Config) {
	// Set mutex profiling fraction if specified in config
	if config.GetMutexProfileRate() != nil {
		runtime.SetMutexProfileFraction(int(config.GetMutexProfileRate().GetValue()))
	}
	// Set block profiling rate if specified in config
	if config.GetBlockProfileRate() != nil {
		runtime.SetBlockProfileRate(int(config.GetBlockProfileRate().GetValue()))
	}
	// Ensure profiler is initialized only once
	once.Do(func() {
		var err error
		// Initialize the pyroscope profiler with provided configuration
		profiler, err = pyroscope.Start(pyroscope.Config{
			ApplicationName:   config.GetAppName().GetValue(),           // Application name to identify the service in pyroscope UI
			Tags:              config.GetTags(),                         // Tags to attach to the profiling data
			ServerAddress:     config.GetServerAddr().GetValue(),        // Address of the pyroscope server
			BasicAuthUser:     config.GetBasicAuthUser().GetValue(),     // HTTP basic auth username
			BasicAuthPassword: config.GetBasicAuthPassword().GetValue(), // HTTP basic auth password
			TenantID:          config.GetTenantId().GetValue(),          // Tenant ID for multi-tenant setups
			UploadRate:        config.GetUploadRate().AsDuration(),      // How often to upload profiling data
			Logger:            pyroscope.StandardLogger,                 // Use standard logger for pyroscope logs
			ProfileTypes: []pyroscope.ProfileType{
				pyroscope.ProfileCPU,           // CPU profiling
				pyroscope.ProfileAllocObjects,  // Allocated objects profiling
				pyroscope.ProfileAllocSpace,    // Allocated space profiling
				pyroscope.ProfileInuseObjects,  // In-use objects profiling
				pyroscope.ProfileInuseSpace,    // In-use space profiling
				pyroscope.ProfileGoroutines,    // Goroutine profiling
				pyroscope.ProfileMutexCount,    // Mutex count profiling
				pyroscope.ProfileMutexDuration, // Mutex wait duration profiling
				pyroscope.ProfileBlockCount,    // Block count profiling
				pyroscope.ProfileBlockDuration, // Block wait duration profiling
			},
			// Disable automatic GC runs between heap profiles if configured
			DisableGCRuns: config.GetDisableGcRuns().GetValue(),
			// Custom HTTP headers to send with requests to pyroscope server
			HTTPHeaders: config.GetHttpHeaders(),
		})
		// Panic if profiler fails to start, as profiling is critical for observability
		if err != nil {
			panic(fmt.Errorf("failed to start profiler, %w", err))
		}
	})
}

// Stop stops the pyroscope profiler if it's currently running
// It safely shuts down the profiler and uploads any remaining profiling data
// If the profiler was never started, this function does nothing
func Stop() {
	if profiler == nil {
		return
	}
	// Stop the profiler and handle any errors during shutdown
	if err := profiler.Stop(); err != nil {
		panic(fmt.Errorf("failed to stop profiler, %w", err))
	}
}
