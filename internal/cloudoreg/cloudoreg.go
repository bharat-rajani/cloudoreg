package cloudoreg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/bharat-rajani/cloudoreg/internal/config"
	"github.com/bharat-rajani/cloudoreg/internal/rhsmapi"
	"github.com/bharat-rajani/cloudoreg/internal/service"
	"github.com/bharat-rajani/cloudoreg/internal/sources"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

const (
	HeaderRHIdentity             = "x-rh-identity"
	HeaderRHSourcesAccountNumber = "x-rh-sources-account-number"
	HeaderRHSourcesOrgID         = "x-rh-sources-org-id"
	HeaderEventType              = "event_type"
	SOURCES_RESOURCE_TYPE        = "Application"
)

var (
	AvailabilityStatusMap = map[bool]string{
		true:  "available",
		false: "unavailable",
	}
)

type Cloudoreg struct {
	sourceEventConsumer  *kafka.Consumer
	sourceStatusProducer *kafka.Producer
	sourceEventTopic     string
	sourceStatusTopic    string
	awsService           service.AWSService
	rhsmAPIService       rhsmapi.Broker
	sourcesService       sources.SourcesService
}

func NewCloudoreg(ctx context.Context, config *config.Config) (*Cloudoreg, error) {
	c, err := kafka.NewConsumer(config.Kafka.SourceEventConsumer)
	if err != nil {
		return nil, fmt.Errorf("cannot create kafka consumer: %w", err)
	}
	go LogKafka(c.Logs())
	p, err := kafka.NewProducer(config.Kafka.SourceStatusProducer)
	if err != nil {
		return nil, fmt.Errorf("cannot create kafka producer: %w", err)
	}
	go LogKafka(p.Logs())

	appSession, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("unable to establish aws session %w", err)
	}

	rhsmAPIService := rhsmapi.NewBroker(config.RHSMService.URL)
	sourcesServiceImpl := sources.New(http.DefaultClient, config.SourcesService.URL, config.SourcesService.PSK)
	return &Cloudoreg{
		sourceEventConsumer:  c,
		sourceStatusProducer: p,
		sourceEventTopic:     config.Kafka.SourceEventTopic,
		sourceStatusTopic:    config.Kafka.SourceStatusTopic,
		awsService:           &service.AWSServiceImpl{AppSession: appSession},
		rhsmAPIService:       *rhsmAPIService,
		sourcesService:       &sourcesServiceImpl,
	}, nil
}

func DefaultCloudoregConfig() *config.Config {
	return &config.Config{
		Kafka: config.KafkaConfig{
			BootstrapServers:  "localhost",
			SourceEventTopic:  "platform.sources.event-stream",
			SourceStatusTopic: "platform.sources.status",
			SourceEventConsumer: &kafka.ConfigMap{
				"bootstrap.servers":      "localhost",
				"group.id":               "cloudoreg.1",
				"auto.offset.reset":      "earliest",
				"enable.auto.commit":     false,
				"go.logs.channel.enable": true,
			},
			SourceStatusProducer: &kafka.ConfigMap{
				"bootstrap.servers":      "localhost",
				"go.logs.channel.enable": true,
			},
		},
		RHSMService: config.RHSMService{URL: "http://localhost:3344/v1"},
		SourcesService: config.SourcesService{
			URL: "http://localhost:4000/api/sources/v3.1",
			PSK: "",
		},
	}
}

func LogKafka(logChan chan kafka.LogEvent) {

	for {
		select {
		case logEvent, ok := <-logChan:
			if !ok {
				return
			}
			fmt.Println(logEvent.String())
		}
	}
}

func (c *Cloudoreg) Run(ctx context.Context) {
	c.ConsumeAndProcess(ctx)
}

func (c *Cloudoreg) ConsumeAndProcess(ctx context.Context) {
	err := c.sourceEventConsumer.Subscribe(c.sourceEventTopic, nil)
	if err != nil {
		log.Printf("error while subscribing to source topic: %e\n", err)
		return
	}
	var ops uint64
	run := true
	fmt.Println("Cloudoreg is listening to souces topic")
	for run == true {
		select {
		case <-ctx.Done():
			log.Println("context done received, exiting ")
			if err = c.sourceEventConsumer.Close(); err != nil {
				log.Println(err)
			}
			run = false
		default:
			ev := c.sourceEventConsumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				c.ProcessSourceEventMessage(ctx, ev.(*kafka.Message))
				atomic.AddUint64(&ops, 1)
				// ensuring at-least once processing.
				topicPartitions, err := c.sourceEventConsumer.CommitMessage(e)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error committing offset after message %s %e:\n",
						e.TopicPartition, err)
				}
				log.Printf("Message processed %v\n", topicPartitions)

			case kafka.Error:
				fmt.Fprintf(os.Stderr, "Error: %v: %v\n", e.Code(), e)
			default:
				fmt.Printf("Processed %d\n", ops)
			}
		}
	}
}

func (c *Cloudoreg) VerifyAccountAccess(awsARN arn.ARN) (bool, error) {
	log.Println("Validating ARN")
	verified := false
	credentials, err := c.awsService.AssumeRole(awsARN)
	if err != nil {
		return verified, err
	}
	customerSession, err := c.awsService.CreateCustomerSession(credentials)
	if err != nil {
		return verified, err
	}
	identity, err := c.awsService.GetCallerIdentity(customerSession)
	if err != nil {
		return verified, err
	}
	identityRoot, err := c.awsService.GetCallerIdentity()
	if err != nil {
		return verified, err
	}

	verified = true
	if identityRoot.Account == identity.Account {
		fmt.Println("WARNING! you are using same account for creation as well as validation of arn")
	} else {
		log.Println("ARN is validated")
	}

	return verified, nil
}

// ProcessApplicationAuthenticationCreate validates the source and establishes the trust at rhsm-api cloudaccess side
func (c *Cloudoreg) ProcessApplicationAuthenticationCreate(ctx context.Context,
	appAuth *ApplicationAuthenticationCreate, ebsAccount, orgID string) (bool, error) {

	validAppTypeID, err := c.sourcesService.GetApplicationTypeIDs(ctx)
	if err != nil {
		log.Println("Error while fetching application type id from sources", err)
	}

	log.Printf("Fetching Application with the given id %s\n", appAuth.ApplicationID.String())
	app, err := c.sourcesService.GetApplicationByID(appAuth.ApplicationID.String(), ebsAccount)
	if err != nil {
		//here we will wait for all the retry to happen
		log.Fatal(err)
		return false, err
	}

	if validAppTypeID != app.ApplicationTypeID {
		return false, nil
	}

	log.Printf("Fetching Authentication with the given id %s\n", appAuth.ApplicationID.String())
	auth, err := c.sourcesService.GetAuthenticationByID(appAuth.AuthenticationID.String(), ebsAccount)
	if err != nil {
		//here we will wait for all the retry to happen
		log.Fatal(err)
		return false, err
	}

	parsedArn, err := arn.Parse(auth.Username)
	if err != nil {
		return false, err
	}
	verified, err := c.VerifyAccountAccess(parsedArn)
	if err != nil {
		return false, err
	}

	if !verified {
		return false, fmt.Errorf("arn could not be verified, err: %w", err)
	}

	err = c.rhsmAPIService.CreateAccountWithAutoReg(rhsmapi.AccountDetails{
		Provider:          "AWS",
		AccountID:         orgID,
		ProviderAccountID: parsedArn.AccountID,
		SourceID:          app.SourceID,
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Cloudoreg) Produce(applicationId string, available bool, err error, headers []kafka.Header) error {
	deliveryChan := make(chan kafka.Event)
	var errString string
	if err != nil {
		errString = err.Error()
	}
	data := map[string]interface{}{
		"resource_type": SOURCES_RESOURCE_TYPE,
		"resource_id":   applicationId,
		"status":        AvailabilityStatusMap[available],
		"error":         errString,
	}
	log.Printf("Source's application availability status %s", AvailabilityStatusMap[available])
	log.Println("Notifying sources about application status by producing kafka message")
	dataBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	headers = append(headers, kafka.Header{
		Key:   "event_type",
		Value: []byte("availability_status"),
	})

	err = c.sourceStatusProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &c.sourceStatusTopic, Partition: kafka.PartitionAny},
		Value:          dataBytes,
		Headers:        headers,
	}, deliveryChan)

	// wait for message to be delivered to kafka topic
	// we process synchronously
	e := <-deliveryChan
	switch e.(type) {
	case *kafka.Message:
		messageAck := e.(*kafka.Message)
		if messageAck.TopicPartition.Error != nil {
			return fmt.Errorf("delivery failed: %w", messageAck.TopicPartition.Error)
		} else {
			log.Printf("Notified sources on topic %s [%d] at offset %v\n",
				*messageAck.TopicPartition.Topic, messageAck.TopicPartition.Partition, messageAck.TopicPartition.Offset)
			return nil
		}
	default:
		return fmt.Errorf("delivery event: %v", e)
	}
}

func (c *Cloudoreg) ProcessSourceEventMessage(ctx context.Context, msg *kafka.Message) {
	headerMap := HeadersAsMap(msg)
	eventType, ok := headerMap[HeaderEventType]
	if !ok {
		return
	}

	switch string(eventType.Value) {
	case "ApplicationAuthentication.create":
		var appAuth ApplicationAuthenticationCreate
		err := json.Unmarshal(msg.Value, &appAuth)
		if err != nil {
			fmt.Println(err)
			return
		}
		ebsAccount, ok := headerMap[HeaderRHSourcesAccountNumber]
		if !ok {
			panic("no account number")
		}
		log.Printf("Received event %s, EBS Account: %s\n", eventType.Value, ebsAccount)
		orgID, ok := headerMap[HeaderRHSourcesOrgID]
		if !ok {
			fmt.Println("Error: no org id number")
		}
		available, err := c.ProcessApplicationAuthenticationCreate(ctx, &appAuth, string(ebsAccount.Value), string(orgID.Value))
		headers := ForwardableMessageHeaders(msg)
		err = c.Produce(appAuth.ApplicationID.String(), available, err, headers)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func ForwardableMessageHeaders(msg *kafka.Message) []kafka.Header {
	var headers []kafka.Header
	for _, h := range msg.Headers {
		h := h
		switch h.Key {
		case HeaderRHIdentity, HeaderRHSourcesOrgID, HeaderRHSourcesAccountNumber:
			headers = append(headers, h)
		}
	}
	return headers
}

func HeadersAsMap(msg *kafka.Message) map[string]kafka.Header {
	headerMap := make(map[string]kafka.Header)
	for _, h := range msg.Headers {
		headerMap[h.Key] = h
	}
	return headerMap
}

type ApplicationAuthenticationCreate struct {
	ID               json.Number `json:"id"`
	CreatedAt        string      `json:"created_at"`
	UpdatedAt        string      `json:"updated_at"`
	PausedAt         string      `json:"paused_at"`
	ApplicationID    json.Number `json:"application_id"`
	AuthenticationID json.Number `json:"authentication_id"`
	Tenant           string      `json:"tenant"`
}
