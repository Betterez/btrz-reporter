package main

import (
	"btrzaws"
	"flag"
	"fmt"
	"log"
	"os"
	"reporter"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func main() {
	analizeCommandLine()
}

const (
	opFileName     = "reporter.txt"
	currentVersion = "1.0.0.7"
)

func analizeCommandLine() {
	currentCommand := flag.String("show", "", "shows current version")
	flag.Parse()
	if *currentCommand == "" {
		isLogginEnalbed()
		reportInstanceStats()
	} else if *currentCommand == "version" {
		fmt.Println("version " + currentVersion)
	} else if *currentCommand == "memory" {
		usage, err := reporter.LoadMemoryValue()
		if err != nil {
			fmt.Println("Error!", err)
			return
		}
		fmt.Printf("\n\nMemory results:\n\tTotal %f\n\tFree %f\n\t%% Occupied %f\n\tOS Version %f\n",
			usage.GetTotalMemory(), usage.GetFreeMemory(), usage.GetUsedMemoryPercentage(), usage.GetOSVersion())
	}
}

func isLogginEnalbed() {
	if _, err := os.Stat(opFileName); os.IsNotExist(err) {
		time.Sleep(time.Minute * 10)
		file, _ := os.OpenFile(opFileName, os.O_CREATE+os.O_RDWR, 0755)
		defer file.Close()
	}
}

func getMetricName(session *session.Session) (metricName string) {
	metricName = os.Getenv("AWS_METRICS")
	if metricName == "" {
		bzInstance, _ := btrzaws.GetSelfInstance(session)
		if bzInstance != nil {
			metricName = bzInstance.InstanceName + "-memory"
		}
	}
	if metricName == "" {
		metricName, _ = btrzaws.GetAwsInstanceID()
	}
	if metricName == "" {
		metricName = "localhost Memory"
	}
	return metricName
}

func reportInstanceStats() {
	session, err := btrzaws.GetAWSSession()
	if err != nil {
		os.Exit(1)
	}
	cloudwatchClient := cloudwatch.New(session)
	if cloudwatchClient == nil {
		fmt.Println("can't create cloud watch client")
	}
	totalErrors := 0
	for {
		if totalErrors > 5 {
			log.Println("Too many errors, exiting")
			os.Exit(1)
		}
		usage, err := reporter.LoadMemoryValue()
		if err != nil {
			totalErrors++
		}
		_, err = cloudwatchClient.PutMetricData(&cloudwatch.PutMetricDataInput{
			Namespace: aws.String(getMetricName(session)),
			MetricData: []*cloudwatch.MetricDatum{
				{
					Dimensions: []*cloudwatch.Dimension{
						{Name: aws.String("Ram KB"), Value: aws.String("Available memory")}},
					MetricName: aws.String("Free Memory"),
					StatisticValues: &cloudwatch.StatisticSet{
						Maximum:     aws.Float64(usage.GetFreeMemory()),
						Minimum:     aws.Float64(usage.GetFreeMemory()),
						SampleCount: aws.Float64(1),
						Sum:         aws.Float64(usage.GetFreeMemory()),
					},
				},
				{
					Dimensions: []*cloudwatch.Dimension{
						{Name: aws.String("Ram KB"), Value: aws.String("Used Memory")}},
					MetricName: aws.String("Used Memory"),
					StatisticValues: &cloudwatch.StatisticSet{
						Maximum:     aws.Float64(usage.GetUsedMemory()),
						Minimum:     aws.Float64(usage.GetUsedMemory()),
						SampleCount: aws.Float64(1),
						Sum:         aws.Float64(usage.GetUsedMemory()),
					},
				},
				{
					Dimensions: []*cloudwatch.Dimension{
						{Name: aws.String("Ram Percentage"), Value: aws.String("Used Memory Percentage")}},
					MetricName: aws.String("Total Memory"),
					StatisticValues: &cloudwatch.StatisticSet{
						Maximum:     aws.Float64(usage.GetUsedMemoryPercentage()),
						Minimum:     aws.Float64(usage.GetUsedMemoryPercentage()),
						SampleCount: aws.Float64(1),
						Sum:         aws.Float64(usage.GetUsedMemoryPercentage()),
					},
				},
			},
		})
		if err != nil {
			totalErrors++
			fmt.Printf("an error occured, %v\n", err)
		} else {
			totalErrors = 0
		}
		time.Sleep(time.Second * 10)
	}
}
