package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/atvaark/dragons-dogma-server/modules/game"
	"github.com/atvaark/dragons-dogma-server/modules/network"
)

type DragonAPIConfig struct {
	Port       int
	ServerHost string
	ServerPort int
	User       string
	UserToken  []byte
}

type DragonAPI struct {
	handler dragonAPIHandler
}

func NewDragonAPI(cfg DragonAPIConfig) *DragonAPI {
	return &DragonAPI{
		handler: dragonAPIHandler{
			cfg: cfg,
			cache: responseCache{
				cacheDuration: 1 * time.Minute,
			},
		},
	}
}

func (d *DragonAPI) ListenAndServe() error {
	log.Printf("[API] Starting\n")

	log.Printf("[API] Testing connection to the game server\n")
	_, err := d.handler.fetchResponse()
	if err != nil {
		return err
	}
	log.Printf("[API] Connection to the game server OK\n")

	mux := http.NewServeMux()
	mux.HandleFunc("/", d.handler.handle)
	addr := fmt.Sprintf(":%d", d.handler.cfg.Port)
	log.Printf("[API] Listening on %s\n", addr)
	err = http.ListenAndServe(addr, mux)
	if err != nil {
		return err
	}
	return nil
}

type dragonAPIHandler struct {
	cfg   DragonAPIConfig
	cache responseCache
}

func (h *dragonAPIHandler) handle(w http.ResponseWriter, r *http.Request) {
	dragonResponse, err := h.fetchResponse()
	if err != nil {
		const getError = "dragon status couldn't be determined"
		log.Printf("%s: %v", getError, err)
		http.Error(w, getError, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(dragonResponse)
}

func (h *dragonAPIHandler) fetchResponse() (*dragonResponse, error) {
	response, err := h.cache.GetResponse()
	if err == nil {
		return response, nil
	}

	response, err = h.cache.UpdateResult(h.fetchNewResponse)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (h *dragonAPIHandler) fetchNewResponse() (*dragonResponse, error) {
	dragon, err := h.getDragon()
	if err != nil {
		return nil, err
	}
	return mapToResponse(dragon), nil
}

func (h *dragonAPIHandler) getDragon() (*game.OnlineUrDragon, error) {
	client := network.NewClient(network.ClientConfig{
		Host:      h.cfg.ServerHost,
		Port:      h.cfg.ServerPort,
		User:      h.cfg.User,
		UserToken: h.cfg.UserToken,
	})

	err := client.Connect()
	if err != nil {
		return nil, err
	}

	dragon, err := client.GetOnlineUrDragon()
	if err != nil {
		return nil, err
	}

	err = client.Disconnect()
	if err != nil {
		return nil, err
	}

	return dragon, nil
}

type dragonResponse struct {
	Generation    int
	SpawnTime     *time.Time
	Defense       int
	FightCount    int
	InGracePeriod bool
	KillTime      *time.Time
	KillCount     int

	Health       int
	HealthTotal  int
	HeartsAlive  int
	HeartsTotal  int
	HeartsHealth float64 // 0 - 11.0

	PawnUserIDs []uint64
}

func mapToResponse(dragon *game.OnlineUrDragon) *dragonResponse {
	response := dragonResponse{
		Generation:    int(dragon.Generation),
		SpawnTime:     dragon.SpawnTime,
		Defense:       int(dragon.Defense),
		FightCount:    int(dragon.FightCount),
		InGracePeriod: dragon.KillTime != nil,
		KillTime:      dragon.KillTime,
		KillCount:     int(dragon.KillCount),
		PawnUserIDs:   dragon.PawnUserIDs[:],
	}

	for _, heart := range dragon.Hearts {
		response.Health += int(heart.Health)
		response.HealthTotal += int(heart.MaxHealth)
		if heart.Health > 0 {
			response.HeartsAlive++
		}
		response.HeartsTotal++
	}

	if response.Health > 0 && response.HealthTotal > 0 {
		response.HeartsHealth = float64(int((float64(response.Health)/float64(response.HealthTotal)*11.0)*10)) / 10
	}

	return &response
}

type responseCache struct {
	response          *dragonResponse
	responseFetchTime time.Time
	mutex             sync.RWMutex
	cacheDuration     time.Duration
}

func (c *responseCache) GetResponse() (*dragonResponse, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.getResponseInternal()
}

func (c *responseCache) UpdateResult(fetchNewFunc func() (*dragonResponse, error)) (*dragonResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	response, err := c.getResponseInternal()
	if err == nil {
		return response, nil
	}

	response, err = fetchNewFunc()
	if err != nil {
		return nil, err
	}

	c.response = response
	c.responseFetchTime = time.Now()

	return response, nil
}

func (c *responseCache) getResponseInternal() (*dragonResponse, error) {
	if c.isCacheValid() {
		response := c.response
		if response == nil {
			return nil, errors.New("invalid cached result")
		}

		return response, nil
	}

	return nil, errors.New("no result cached")
}

func (c *responseCache) isCacheValid() bool {
	return c.responseFetchTime.After(time.Now().Add(-c.cacheDuration))
}
