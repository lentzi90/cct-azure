package cct_azure

import (
	"flag"
	"log"
	"time"

	"github.com/Azure/go-autorest/autorest"
)

var (
	subscriptionID = flag.String("subscription-id", "", "The ID of the subscription.")
	addr           = flag.String("listen-address", ":8080", "The address to listen on for (scraping) HTTP requests.")
)

var authorizer autorest.Authorizer

func main() {
	flag.Parse()
	if *subscriptionID == "" {
		log.Fatal("You must provide a subscription id by using the --subscription-id flag.")
	}

	log.Println("Initializing client...")
	client := NewRestClient(*subscriptionID)
	usageExplorer := NewUsageExplorer(client)
	date := time.Date(2018, time.July, 3, 00, 0, 0, 0, time.UTC)
	usageExplorer.GetCloudCost(date)
}
