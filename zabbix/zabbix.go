package zabbix

import (
	"encoding/json"
	"fmt"
	"github.com/akomic/go-zabbix-proto/client"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strings"
	"unicode"
)

const (
	namespace = "zabbix"
)

// Zabbix structure.
type Zabbix struct {
	Client *client.Client
}

// NewZabbix constructor
func NewZabbix(host string, port int) (p *Zabbix) {
	p = &Zabbix{
		Client: client.NewClient(host, port),
	}
	return
}

// GetMetrics from Zabbix.
func (zabbix *Zabbix) GetMetrics() (response *ZabbixResponse, err error) {
	packet := zabbix.NewStatsPacket(`zabbix.stats`)

	var res []byte
	res, err = zabbix.Client.Send(packet)
	if err != nil {
		return
	}

	response, err = NewZabbixResponse(res)
	if err != nil {
		return
	}

	if response.Response != `success` {
		err = fmt.Errorf("Error sending heartbeat: %s", response.Response)
	}
	return
}

// StatsPacket structure.
type StatsPacket struct {
	Request string `json:"request"`
}

// NewStatsPacket constructor.
func (zabbix *Zabbix) NewStatsPacket(request string) *client.Packet {
	ap := &StatsPacket{Request: request}
	jsonData, _ := json.Marshal(ap)
	packet := &client.Packet{Request: request, Data: jsonData}
	return packet
}

// ZabbixResponse structure.
type ZabbixResponse struct {
	Response string `json:"response"`
	Data     map[string]interface{}
}

// NewZabbixResponse constructor.
func NewZabbixResponse(data []uint8) (r *ZabbixResponse, err error) {
	if len(data) < 13 {
		err = fmt.Errorf("NewZabbixResponse Input data to short")
		return
	}
	jsonData := data[13:]
	r = &ZabbixResponse{Response: string(jsonData)}
	err = json.Unmarshal(jsonData, r)
	if err != nil {
		err = fmt.Errorf("Error decoding response: %v", err)
		if e, ok := err.(*json.SyntaxError); ok {
			err = fmt.Errorf("%s ; Syntax error at byte offset %d", err, e.Offset)
		}
		return
	}
	return
}

// Describe describe metrics
func (zabbix *Zabbix) Describe(ch chan<- *prometheus.Desc) {

	metricCh := make(chan prometheus.Metric)
	doneCh := make(chan struct{})

	go func() {
		for m := range metricCh {
			ch <- m.Desc()
		}
		close(doneCh)
	}()

	zabbix.Collect(metricCh)
	close(metricCh)
	<-doneCh
}

// Collect metrics
func (zabbix *Zabbix) Collect(ch chan<- prometheus.Metric) {
	upValue := 1

	if err := zabbix.collect(ch); err != nil {
		log.Printf("Error scraping zabbix: %s", err)
		upValue = 0
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Was the last query of Zabbix successful.",
			nil, nil,
		),
		prometheus.GaugeValue, float64(upValue),
	)
}

func (zabbix *Zabbix) collect(ch chan<- prometheus.Metric) error {
	metrics, err := zabbix.GetMetrics()
	if err != nil {
		return fmt.Errorf("Error scraping zabbix: %v", err)
	}
	getMetricRecursive(metrics.Data, ch, "")
	return nil
}

func getMetricRecursive(metrics map[string]interface{}, ch chan<- prometheus.Metric, prefix string) {
	for key, value := range metrics {
		name := prefix + key
		switch value.(type) {
		case float64:
			newMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      metricName(name),
				//Help:      "Number of " + name + " currently processed",
			}, []string{}).WithLabelValues()
			newMetric.Set(value.(float64))
			newMetric.Collect(ch)
		case string:
			if key == "version" {
				newMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      metricName("info"),
					Help:      "Info abount instance",
				}, []string{"version"}).WithLabelValues(value.(string))
				newMetric.Set(1)
				newMetric.Collect(ch)
			}
		case []interface{}:
			parseSlice(ch, key, value.([]interface{}))
		case map[string]interface{}:
			//log.Printf("other %v",value)
			getMetricRecursive(value.(map[string]interface{}), ch, name+"_")
		}
	}
}

// parses slice section all strings values become labels
func parseSlice(ch chan<- prometheus.Metric, category string, items []interface{}) {

	for _, item := range items {

		labels := make(map[string]string)
		labelsNames := make([]string, 0)
		if p, ok := item.(map[string]interface{}); ok {
			// Get all strings as labels
			for key, value := range p {
				if v, ok := value.(string); ok {
					labels[key] = v
					labelsNames = append(labelsNames, key)
				}
			}
			for key, value := range p {
				var floatMetric float64 = 0
				name := category + "_" + key
				switch value.(type) {
				case float64:
					floatMetric = value.(float64)
				case bool:
					if value.(bool) {
						floatMetric = 1
					} else {
						floatMetric = 0
					}
				default:
					continue
				}
				newMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Namespace: namespace,
					Name:      name,
				}, labelsNames).With(labels)
				newMetric.Set(floatMetric)
				newMetric.Collect(ch)
			}
		}
	}
}

func metricName(in string) string {
	out := toSnake(in)
	r := strings.NewReplacer(".", "_", " ", "", "-", "")
	return r.Replace(out)
}

func toSnake(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// check interface
var _ prometheus.Collector = (*Zabbix)(nil)
