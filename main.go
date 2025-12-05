package main

import (
	"os"

	"salesforce-splunk-migration/cmd"
	"salesforce-splunk-migration/utils"
)

func main() {
	// Initialize global logger
	if err := utils.InitializeGlobalLogger("salesforce-splunk-migration", "main", true); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	logger := utils.GetLogger()

	if err := cmd.Execute(); err != nil {
		logger.Error("Migration failed", utils.Err(err))
		os.Exit(1)
	}
}
