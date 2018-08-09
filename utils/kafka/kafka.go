package kafka

import (
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/Shopify/sarama"
	"github.com/krise3k/armada-stats/utils"
)

var ValidTopicSuffixMethods = []string{
	"",
	"measurement",
	"tags",
}

type (
	Kafka struct {
		// Kafka brokers to send metrics to
		Brokers []string
		// Kafka topic
		Topic string
		// Kafka client id
		ClientID string
		// Kafka topic suffix option
		TopicSuffix TopicSuffix
		// Routing Key Tag
		RoutingTag string
		// Compression Codec Tag
		CompressionCodec int
		// RequiredAcks Tag
		RequiredAcks int
		// MaxRetry Tag
		MaxRetry int

		producer   sarama.SyncProducer
		serializer serializers.Serializer
	}
	TopicSuffix struct {
		Method    string
		Keys      []string
		Separator string
	}
)

func ValidateTopicSuffixMethod(method string) error {
	for _, validMethod := range ValidTopicSuffixMethods {
		if method == validMethod {
			return nil
		}
	}
	return fmt.Errorf("Unknown topic suffix method provided: %s", method)
}

func (k *Kafka) GetTopicName(metric telegraf.Metric) string {
	var topicName string
	switch k.TopicSuffix.Method {
	case "measurement":
		topicName = k.Topic + k.TopicSuffix.Separator + metric.Name()
	case "tags":
		var topicNameComponents []string
		topicNameComponents = append(topicNameComponents, k.Topic)
		for _, tag := range k.TopicSuffix.Keys {
			tagValue := metric.Tags()[tag]
			if tagValue != "" {
				topicNameComponents = append(topicNameComponents, tagValue)
			}
		}
		topicName = strings.Join(topicNameComponents, k.TopicSuffix.Separator)
	default:
		topicName = k.Topic
	}
	return topicName
}

func (k *Kafka) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *Kafka) Connect() error {
	err := ValidateTopicSuffixMethod(k.TopicSuffix.Method)
	if err != nil {
		return err
	}
	config := sarama.NewConfig()

	if k.ClientID != "" {
		config.ClientID = k.ClientID
	} else {
		config.ClientID = "Telegraf"
	}

	config.Producer.RequiredAcks = sarama.RequiredAcks(k.RequiredAcks)
	config.Producer.Compression = sarama.CompressionCodec(k.CompressionCodec)
	config.Producer.Retry.Max = k.MaxRetry
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(k.Brokers, config)
	if err != nil {
		return err
	}
	k.producer = producer
	return nil
}

func (k *Kafka) Close() error {
	return k.producer.Close()
}

func (k *Kafka) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := k.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		topicName := k.GetTopicName(metric)

		m := &sarama.ProducerMessage{
			Topic: topicName,
			Value: sarama.ByteEncoder(buf),
		}
		if h, ok := metric.Tags()[k.RoutingTag]; ok {
			m.Key = sarama.StringEncoder(h)
		}

		_, _, err = k.producer.SendMessage(m)

		if err != nil {
			return fmt.Errorf("FAILED to send kafka message: %s\n", err)
		}
	}
	return nil
}

var kafkaClient *Kafka

func initClient() {
	brokers, _ := utils.Config.List("kafka_brokers")
	topic, _ := utils.Config.String("kafka_topic")
	maxRetry, _ := utils.Config.Int("kafka_max_retry")
	var brokersList []string
	for _, broker := range brokers {
		brokersList = append(brokersList, broker.(string))
	}

	var k = &Kafka{Brokers: brokersList, Topic: topic, MaxRetry: maxRetry}
	k.SetSerializer(influx.NewSerializer())
	k.Connect()
	kafkaClient = k

}

func GetKafkaClient() *Kafka {
	if kafkaClient == nil {
		initClient()
	}

	return kafkaClient
}
