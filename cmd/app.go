// Package cmd implements the main application logic
package cmd

import (
	"fmt"
	"time"

	"salesforce-splunk-migration/services"
	"salesforce-splunk-migration/utils"
)

// Execute runs the main migration workflow
func Execute() error {
	fmt.Println("🚀 Starting Salesforce to Splunk Migration...")

	// Load configuration
	config, err := utils.LoadConfig("credentials.json")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	fmt.Println("✅ Configuration loaded and validated")

	// Create Splunk service
	splunkService, err := services.NewSplunkService(config)
	if err != nil {
		return fmt.Errorf("failed to create Splunk service: %w", err)
	}

	// Step 1: Authenticate
	fmt.Println("\n🔐 Step 1: Authenticating with Splunk...")
	if err := splunkService.Authenticate(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	fmt.Println("✅ Authentication successful")

	// Step 2: Create Index
	fmt.Println("\n📊 Step 2: Creating Splunk index...")
	if err := splunkService.CreateIndex(config.Splunk.IndexName); err != nil {
		return fmt.Errorf("failed to create index %s: %w", config.Splunk.IndexName, err)
	}
	fmt.Printf("✅ Index created: %s\n", config.Splunk.IndexName)

	// Step 2a: Verify Index Exists
	// fmt.Printf("🔍 Verifying index '%s' exists...\n", config.Splunk.IndexName)
	// if err := splunkService.VerifyIndexExists(config.Splunk.IndexName); err != nil {
	// 	return fmt.Errorf("index verification failed: %w. Cannot proceed with data inputs creation", err)
	// }
	// fmt.Printf("✅ Index '%s' verified successfully\n", config.Splunk.IndexName)

	// Step 3: Create Salesforce Account
	fmt.Println("\n🔗 Step 3: Creating Salesforce account in Splunk...")
	if err := splunkService.CreateSalesforceAccount(); err != nil {
		return fmt.Errorf("failed to create Salesforce account: %w", err)
	}
	fmt.Println("✅ Salesforce account created")

	// Step 3a: Verify Account Exists
	// fmt.Printf("🔍 Verifying Salesforce account '%s' exists...\n", config.Salesforce.AccountName)
	// if err := splunkService.VerifyAccountExists(config.Salesforce.AccountName); err != nil {
	// 	return fmt.Errorf("account verification failed: %w. Cannot proceed with data inputs creation", err)
	// }
	// fmt.Printf("✅ Salesforce account '%s' verified successfully\n", config.Salesforce.AccountName)

	// Step 4: Create Data Inputs (only after index and account are verified)
	dataInputs, err := config.GetDataInputs()
	if err != nil {
		return fmt.Errorf("failed to load data inputs: %w", err)
	}

	fmt.Printf("\n📥 Step 4: Creating %d data inputs...\n", len(dataInputs))

	if len(dataInputs) == 0 {
		fmt.Println("⚠️  No data inputs configured. Skipping...")
	} else {
		successCount := 0
		failedCount := 0

		for i, input := range dataInputs {
			fmt.Printf("\n[%d/%d] Creating data input: %s (Object: %s)\n", i+1, len(dataInputs), input.Name, input.Object)

			if err := splunkService.CreateDataInput(&input); err != nil {
				fmt.Printf("  ❌ Failed: %v\n", err)
				failedCount++
			} else {
				fmt.Printf("  ✅ Created successfully\n")
				successCount++
			}

			// Small delay between data input creations
			if i < len(dataInputs)-1 {
				time.Sleep(500 * time.Millisecond)
			}
		}

		fmt.Printf("\n📊 Summary: %d/%d data inputs created successfully", successCount, len(dataInputs))
		if failedCount > 0 {
			fmt.Printf(", %d failed\n", failedCount)
			return fmt.Errorf("%d data inputs failed to create", failedCount)
		} else {
			fmt.Println()
		}
	}

	// Step 5: Verify only the data inputs created by this code
	fmt.Println("\n🔍 Step 5: Verifying created data inputs...")
	if existingInputs, err := splunkService.ListDataInputs(); err == nil {
		// Create a map of created input names for quick lookup
		createdInputNames := make(map[string]bool)
		for _, input := range dataInputs {
			createdInputNames[input.Name] = true
		}

		// Filter to show only inputs created by this code
		var ourInputs []string
		for _, name := range existingInputs {
			if createdInputNames[name] {
				ourInputs = append(ourInputs, name)
			}
		}

		if len(ourInputs) > 0 {
			fmt.Printf("✅ Verified %d data inputs created by this code:\n", len(ourInputs))
			for _, name := range ourInputs {
				fmt.Printf("  - %s\n", name)
			}
		} else {
			fmt.Println("⚠️  Warning: None of the configured data inputs were found in Splunk")
		}
	} else {
		fmt.Printf("⚠️  Warning: Could not list data inputs: %v\n", err)
	}

	fmt.Println("\n🎉 Migration completed successfully!")
	return nil
}
