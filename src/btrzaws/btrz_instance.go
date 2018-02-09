package btrzaws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// BetterezInstance - aws representation, for betterez
type BetterezInstance struct {
	Environment            string
	Repository             string
	PrivateIPAddress       string
	PublicIPAddress        string
	BuildNumber            int
	KeyName                string
	InstanceName           string
	InstanceID             string
	PathName               string
	ServiceStatus          string
	ServiceStatusErrorCode string
	FaultsCount            int
	StatusCheck            time.Time
}

const (
	// ConnectionTimeout - waiting time in which healthchceck should be back
	ConnectionTimeout = time.Duration(5 * time.Second)
	// AwsCheckAddress -url to check aws instance id
	AwsCheckAddress = "http://169.254.169.254/latest/meta-data/instance-id"
)

// GetAwsInstanceID - gets current aws instance id (if there is one)
func GetAwsInstanceID() (string, error) {
	httpClient := &http.Client{Timeout: time.Second * 3}
	resp, err := httpClient.Get(AwsCheckAddress)
	if err != nil {
		if strings.Index(err.Error(), "no route to host") != -1 {
			return "localhost", nil
		}
		return "", err
	}
	instanceIDData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	instanceID := string(instanceIDData)
	return instanceID, nil
}

// GetSelfInstance - return self BetterezInstance
func GetSelfInstance(awsSession *session.Session) (instance *BetterezInstance, err error) {
	instanceID, err := GetAwsInstanceID()
	if err != nil {
		return nil, err
	}
	instance, err = GetInstanceInfoFromInstanceID(awsSession, instanceID)
	return instance, err
}

// GetInstanceInfoFromInstanceID - return BetterezInstance for given instanceId
func GetInstanceInfoFromInstanceID(awsSession *session.Session, instanceID string) (instance *BetterezInstance, err error) {
	ec2Servicve := ec2.New(awsSession)
	response, err := ec2Servicve.DescribeInstances(&ec2.DescribeInstancesInput{
		DryRun:      aws.Bool(false),
		InstanceIds: []*string{aws.String(instanceID)},
	})
	if err != nil {
		return nil, err
	}
	if len(response.Reservations) == 0 {
		return nil, nil
	}
	for _, reservation := range response.Reservations {
		if len(reservation.Instances) == 0 {
			continue
		} else {
			awsInstance := reservation.Instances[0]
			instance = LoadFromAWSInstance(awsInstance)
			break
		}

	}
	return instance, nil
}

// LoadFromAWSInstance - returns new BetterezInstance or an error
func LoadFromAWSInstance(instance *ec2.Instance) *BetterezInstance {
	result := &BetterezInstance{
		Environment:  GetTagValue(instance, "Environment"),
		Repository:   GetTagValue(instance, "Repository"),
		PathName:     GetTagValue(instance, "Path-Name"),
		InstanceName: GetTagValue(instance, "Name"),
		InstanceID:   *instance.InstanceId,
		KeyName:      *instance.KeyName,
	}
	if instance.PublicIpAddress != nil {
		result.PublicIPAddress = *instance.PublicIpAddress
	}

	if instance.PrivateIpAddress != nil {
		result.PrivateIPAddress = *instance.PrivateIpAddress
	}
	buildNumber, err := strconv.Atoi(GetTagValue(instance, "Build-Number"))
	if err != nil {
		result.BuildNumber = 0
	} else {
		result.BuildNumber = buildNumber
	}
	return result
}

// GetHealthCheckString - Creates the healthcheck string based on the service name and address
func (instance *BetterezInstance) GetHealthCheckString() string {
	port := 3000
	var testURL string
	var testIPAddress string
	if instance.PublicIPAddress != "" {
		testIPAddress = instance.PublicIPAddress
	} else {
		testIPAddress = instance.PrivateIPAddress
	}
	if instance.Repository == "connex2" {
		port = 22000
		testURL = fmt.Sprintf("http://%s:%d/healthcheck", testIPAddress, port)
	} else {
		testURL = fmt.Sprintf("http://%s:%d/%s/healthcheck", testIPAddress, port, instance.PathName)
	}
	return testURL
}

// CheckIsnstanceHealth - checks instance health
func (instance *BetterezInstance) CheckIsnstanceHealth() (bool, error) {
	if instance == nil || instance.PrivateIPAddress == "" {
		return true, nil
	}
	httpClient := http.Client{Timeout: ConnectionTimeout}
	resp, err := httpClient.Get(instance.GetHealthCheckString())
	instance.StatusCheck = time.Now()
	if err != nil {
		instance.ServiceStatus = "offline"
		instance.ServiceStatusErrorCode = fmt.Sprintf("%v", err)
		//log.Printf("Error %v healthcheck instance %s", err, instance.InstanceID)
		return false, err
	}
	defer resp.Body.Close()
	//log.Print("checking ", instance.Repository, "...")
	if resp.StatusCode == 200 {
		instance.ServiceStatus = "online"
		instance.ServiceStatusErrorCode = ""
		return true, nil
	}
	return false, nil
}
