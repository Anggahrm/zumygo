package commands

import (
	"zumygo/helpers"
	"zumygo/libs"
)

func performance(conn *libs.IClient, m *libs.IMessage) bool {
	// Get performance monitor
	monitor := helpers.GetPerformanceMonitor()
	
	// Get performance report
	report := monitor.GetPerformanceReport()
	
	// Send the report
	m.Reply(report)
	return true
}

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:     "performance",
		As:       []string{"performance", "perf", "stats"},
		Tags:     "main",
		IsPrefix: true,
		Execute:  performance,
	})
} 