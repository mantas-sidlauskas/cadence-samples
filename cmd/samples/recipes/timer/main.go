package main

import (
	"flag"
	"sync"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.WorkerMetricScope,
		Logger:       h.Logger,
	}
	for i := 0; i < len(UUIDS); i++ {
		go h.StartWorkers(h.Config.DomainName, UUIDS[i], workerOptions)
	}

}

func startWorkflow(h *common.SampleHelper) {
	wg := sync.WaitGroup{}
	for i := 0; i < len(UUIDS); i++ {

		workflowOptions := client.StartWorkflowOptions{
			ID:                              "timer_" + uuid.New(),
			TaskList:                        UUIDS[i],
			ExecutionStartToCloseTimeout:    time.Minute,
			DecisionTaskStartToCloseTimeout: time.Minute,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.StartWorkflow(workflowOptions, sampleTimerWorkflow, time.Minute*5)
		}()
	}
	wg.Wait()

}

func main() {
	var mode string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		h.RegisterWorkflow(sampleTimerWorkflow)
		h.RegisterActivity(orderProcessingActivity)
		h.RegisterActivity(sendEmailActivity)
		startWorkers(&h)

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger":
		startWorkflow(&h)
	}
}
