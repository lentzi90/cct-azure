package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/consumption/mgmt/2018-05-31/consumption"
	"github.com/Azure/azure-sdk-for-go/services/preview/billing/mgmt/2018-03-01-preview/billing"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/shopspring/decimal"
)

// A UsageExplorer can be used to investigate usage cost
type UsageExplorer struct {
	authorizer     autorest.Authorizer
	subscriptionID string
}

// NewUsageExplorer initializes a UsageExplorer
func NewUsageExplorer(subscriptionID string) UsageExplorer {
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	return UsageExplorer{authorizer: authorizer, subscriptionID: subscriptionID}
}

// PrintCurrentUsage prints the usage for the current billing period
func (e *UsageExplorer) PrintCurrentUsage() {
	periods := e.getPeriodsIterator()
	usageIterator := e.getUsageIterator(*periods.Value().Name)
	fmt.Println("Pretax cost Currency, Usage start - Usage end, Provider")
	fmt.Println("----------------------------------------------------------")
	// For all values, print some information
	for usageIterator.NotDone() {
		usageDetails := usageIterator.Value()
		instanceID := *usageDetails.InstanceID
		pretaxCost := *usageDetails.PretaxCost
		currency := *usageDetails.Currency
		usageStart := *usageDetails.UsageStart
		usageEnd := *usageDetails.UsageEnd
		// isEstimated := *usageDetails.IsEstimated

		resourceProvider := getProvider(instanceID)
		fmt.Printf("%s %s, %s - %s, %s\n", pretaxCost, currency, usageStart.Format("2006-01-02 15:04"), usageEnd.Format("2006-01-02 15:04"), resourceProvider)
		usageIterator.Next()
	}
}

func (e *UsageExplorer) getPeriodsIterator() billing.PeriodsListResultIterator {
	periodsClient := billing.NewPeriodsClient(e.subscriptionID)
	periodsClient.Authorizer = e.authorizer

	// filter := "billingPeriodEndDate lt 2018-05-30"
	filter := ""

	periods, err := periodsClient.ListComplete(context.Background(), filter, "", nil)
	if err != nil {
		log.Fatal(err)
	}

	return periods
}

func (e *UsageExplorer) getPeriodByDate(date time.Time) billing.Period {
	periodsClient := billing.NewPeriodsClient(e.subscriptionID)
	periodsClient.Authorizer = e.authorizer

	dateStr := date.Format("2006-01-02")
	filter := "billingPeriodEndDate gt " + dateStr

	periods, err := periodsClient.ListComplete(context.Background(), filter, "", nil)
	if err != nil {
		log.Fatal(err)
	}
	// Periods are returned in reverse chronologic order, so we return the first one.
	// This will be the billing period including date
	return periods.Value()
}

func (e *UsageExplorer) getUsageIterator(billingPeriod string) consumption.UsageDetailsListResultIterator {
	usageClient := consumption.NewUsageDetailsClient(e.subscriptionID)
	usageClient.Authorizer = e.authorizer

	expand := ""
	// filter := "properties/usageEnd le '2018-07-02' AND properties/usageEnd ge '2018-06-30'"
	filter := ""
	skiptoken := ""
	var top int32 = 100
	apply := ""
	log.Println("Trying to get list from billing period", billingPeriod)
	result, err := usageClient.ListByBillingPeriodComplete(context.Background(), billingPeriod, expand, filter, apply, skiptoken, &top)

	if err != nil {
		log.Fatal(err)
	}

	return result
}

func (e *UsageExplorer) getUsageByDate(date time.Time) consumption.UsageDetailsListResultIterator {
	usageClient := consumption.NewUsageDetailsClient(e.subscriptionID)
	usageClient.Authorizer = e.authorizer

	billingPeriod := *e.getPeriodByDate(date).Name

	expand := ""
	filter := fmt.Sprintf("properties/usageStart eq '%s'", date.Format("2006-01-02"))
	skiptoken := ""
	var top int32 = 100
	apply := ""
	log.Println("Trying to get list from billing period", billingPeriod)
	result, err := usageClient.ListByBillingPeriodComplete(context.Background(), billingPeriod, expand, filter, apply, skiptoken, &top)

	if err != nil {
		log.Fatal(err)
	}

	return result
}

func getProvider(instanceID string) string {
	// The instance ID is a string like this:
	// /subscriptions/{guid}/resourceGroups/{resource-group-name}/{resource-provider-namespace}/{resource-type}/{subtype}/{resource-name}
	// See: https://docs.microsoft.com/en-us/rest/api/resources/resources/getbyid
	// We extract the provider by splitting on /
	parts := strings.Split(instanceID, "/")
	return strings.Join(parts[6:8], "/")
}

func (e *UsageExplorer) getUsageDetails(date time.Time) {
	usageIterator := e.getUsageByDate(date)
	providers := make(map[string]decimal.Decimal)

	for usageIterator.NotDone() {
		usageDetails := usageIterator.Value()
		instanceID := *usageDetails.InstanceID
		pretaxCost := *usageDetails.PretaxCost
		currency := *usageDetails.Currency
		usageStart := *usageDetails.UsageStart
		usageEnd := *usageDetails.UsageEnd
		// isEstimated := *usageDetails.IsEstimated

		resourceProvider := getProvider(instanceID)
		providers[resourceProvider] = decimal.Sum(providers[resourceProvider], pretaxCost)
		fmt.Printf("%s %s, %s - %s, %s\n", pretaxCost, currency, usageStart.Format("2006-01-02 15:04"), usageEnd.Format("2006-01-02 15:04"), resourceProvider)

		usageIterator.Next()
	}

	fmt.Println(providers)
}
