package heart

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Heartbeat struct {
	Address     string `json:"address"`
	ServiceName string `json:"service_name"`
	TxTimestamp int64  `json:"tx_timestamp"`
	RxTimestamp int64  `json:"rx_timestamp"`
	Ping        int64  `json:"ping"`
}

type Monitor struct {
	sync.Mutex
	c          *time.Ticker
	services   map[string]Service
	heartbeats []Heartbeat
}

func NewMonitor() Monitor {
	return Monitor{
		c:        time.NewTicker(time.Second),
		services: make(map[string]Service),
	}
}

func (m *Monitor) Ping() {
	m.c = time.NewTicker(time.Second)
	for range m.c.C {
		go func(m *Monitor) {
			for _, serv := range m.services {
				err := serv.ping(m)
				if err != nil {
					fmt.Println(err)
				}
			}
		}(m)
	}
}

func (m *Monitor) MonitorHandler(w http.ResponseWriter, _ *http.Request) {
	m.Lock()
	err := json.NewEncoder(w).Encode(m.heartbeats)
	if err != nil {
		fmt.Fprint(w, err)
	}
	defer m.Unlock()
}

// Service
//=============================================================================================
type Service struct {
	Address string
	Name    string
}

func (s *Service) ServiceHandler(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(Heartbeat{
		Address:     s.Address,
		ServiceName: s.Name,
		TxTimestamp: time.Now().Unix(),
	})
	if err != nil {
		fmt.Fprint(w, err)
	}
}

func (s *Service) ping(m *Monitor) error {
	resp, err := http.Get(s.Address)
	if err != nil {
		return err
	}

	h := Heartbeat{
		RxTimestamp: time.Now().Unix(),
	}

	h.Ping = h.RxTimestamp - h.TxTimestamp

	err = json.NewDecoder(resp.Body).Decode(&h)
	if err != nil {
		return err
	}

	m.Lock()
	m.heartbeats = append(m.heartbeats, h)
	defer m.Unlock()

	return nil
}
