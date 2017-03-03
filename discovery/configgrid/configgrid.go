package configgrid

import (
	"golang.org/x/net/context"
	"encoding/json"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
	"net/http"
	"strconv"
	"time"
)

type configGridJSON struct {
	Container   string      `json:"container"`
	Dc          string      `json:"dc"`
	Env         string      `json:"env"`
	Group       string      `json:"group"`
	Hostname    string      `json:"hostname"`
	ID          int         `json:"id"`
	MetricsPort string      `json:"metrics_port"`
	Module      string      `json:"module"`
	Owner       string      `json:"owner"`
	Path        string      `json:"path"`
	Port        string      `json:"port"`
	Post        interface{} `json:"post"`
	Project     string      `json:"project"`
	Runas       string      `json:"runas"`
	Stripe      interface{} `json:"stripe"`
}

type configgrid struct {
	Configs []configGridJSON `json:"configs"`
}

type ConfigGridDiscovery struct {
	URL             string
	Environment     string
	Project         string
	Datacenter      string
	MetricsPort     int
	RefreshInterval time.Duration
}

func getConfigGrid(url string) (configgrid, error) {
	var body configgrid
	resp, err := http.Get(url)
	if err != nil {
		return body, err
	}

	json.NewDecoder(resp.Body).Decode(&body)

	return body, nil
}

func NewDiscovery(conf *config.ConfigGridConfig) *ConfigGridDiscovery {
	return &ConfigGridDiscovery{
		URL:             conf.URL.URL.String(),
		Environment:     conf.Environment,
		Project:         conf.Project,
		Datacenter:      conf.Datacenter,
		MetricsPort:     conf.Port,
		RefreshInterval: conf.RefreshInterval,
	}
}

func (cg *ConfigGridDiscovery) Run(ctx context.Context, ch chan<- []*config.TargetGroup) {
	log.Debug("Starting discovery via config grid")
	ticker := time.NewTicker(cg.RefreshInterval)
	defer ticker.Stop()

	cg.refresh()
	for {
		select {
		case <-ticker.C:
			tg, err := cg.refresh()
			if err != nil {
				log.Error(err)
				continue
			}

			select {
			case ch <- []*config.TargetGroup{tg}:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (cg *ConfigGridDiscovery) refresh() (*config.TargetGroup, error) {
	tg := &config.TargetGroup{
		Source: "CONFIG_GRID_" + cg.Project + "_" + cg.Datacenter + "_" + cg.Environment,
	}
	configs, err := getConfigGrid(cg.URL)
	if err != nil {
		log.Error("Error getting config grid: ", err)
		return tg, err
	}

	for _, conf := range configs.Configs {
		if conf.Project == cg.Project && conf.Env == cg.Environment && conf.Dc == cg.Datacenter {
			log.Debugf("Adding AddressLabel: %s", conf.Hostname + ":" + strconv.Itoa(cg.MetricsPort))
			labels := model.LabelSet{}
			labels[model.AddressLabel] = model.LabelValue(conf.Hostname + ":" + strconv.Itoa(cg.MetricsPort))
			tg.Targets = append(tg.Targets, labels)
		}
	}

	return tg, nil
}
