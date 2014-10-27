package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/cloudwatch"
)

func perror(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	statusURL := fmt.Sprintf("%s/queue/api/json?pretty=true", os.Getenv("JENKINS_URL"))
	req, _ := http.NewRequest("GET", statusURL, nil)
	req.SetBasicAuth(os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	perror(err)

	body, err := ioutil.ReadAll(resp.Body)
	perror(err)

	var tmp map[string][]interface{}
	json.Unmarshal(body, &tmp)

	if _, ok := tmp["items"]; !ok {
		log.Fatalf("Unexpected JSON returned: %v", tmp)
	}

	auth, err := aws.GetAuth("", "", "", time.Now())
	perror(err)

	region, ok := aws.Regions[os.Getenv("AWS_REGION")]
	if !ok {
		log.Fatal("Region info not found. Please provide AWS_REGION env.")
	}

	cw, err := cloudwatch.NewCloudWatch(auth, region.CloudWatchServicepoint)
	perror(err)

	_, err = cw.PutMetricDataNamespace([]cloudwatch.MetricDatum{
		cloudwatch.MetricDatum{
			MetricName: "Queued-Jenkins-Jobs",
			Value:      float64(len(tmp["items"])),
			Unit:       "Count",
		},
	}, "Jenkins")
	perror(err)

}
