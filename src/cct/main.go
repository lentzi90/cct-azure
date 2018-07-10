package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"cct/cct_azure"
	"cct/db_client"
)

var subscriptionID = flag.String("subscription-id", "", "The ID of the subscription.")

func main() {
	fmt.Println("The coolest DB Client V1.0.5")

	dbConfig := db_client.DBClientConfig{
		DBName:   "prometheus",
		Username: "prom",
		Password: "prom",
		Address:  "http://localhost:8086",
	}

	flag.Parse()
	if *subscriptionID == "" {
		log.Fatal("You must provide a subscription id by using the --subscription-id flag.")
	}

	log.Println("Initializing client...")
	client := cct_azure.NewRestClient(*subscriptionID)
	usageExplorer := cct_azure.NewUsageExplorer(client)

	db := db_client.NewDBClient(dbConfig)
	now := time.Date(2017, time.October, 28, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 35; i++ {
		fetchTime := now.AddDate(0, 0, -i)
		fmt.Println("Getting for period", fetchTime)
		test := usageExplorer.GetCloudCost(fetchTime)
		db.AddUsageData(test)
	}

	log.Println("DONE!!!")
}