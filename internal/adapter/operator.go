package adapter

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jgivc/vapp/config"
	"github.com/jgivc/vapp/internal/entity"
)

type APIClient interface {
	Get(ctx context.Context) (io.ReadCloser, error)
}

type httpAPIClient struct {
	noVerify   bool
	apiURL     string
	apiTimeout time.Duration
}

func (c *httpAPIClient) Get(ctx context.Context) (io.ReadCloser, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.noVerify},
	}

	var client = &http.Client{Timeout: c.apiTimeout, Transport: transport}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create http request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

type RawOperators struct {
	Operators []*entity.Operator `json:"operators"`
}

type OperatorRepo struct {
	mux       sync.Mutex
	apiClient APIClient
	operators map[string]*entity.Operator
}

func (r *OperatorRepo) load(ctx context.Context) ([]*entity.Operator, error) {
	resp, err := r.apiClient.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot call web api: %w", err)
	}
	defer resp.Close()

	var operators RawOperators

	if err = json.NewDecoder(resp).Decode(&operators); err != nil {
		return nil, fmt.Errorf("cannot load operators: %w", err)
	}

	return operators.Operators, nil
}

func (r *OperatorRepo) GetOperators(ctx context.Context) ([]*entity.Operator, error) {
	ops, err := r.load(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load operators: %w", err)
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	m := make(map[string]struct{})
	for i, rOp := range ops {
		m[rOp.Number] = struct{}{}
		if lOp, exists := r.operators[rOp.Number]; exists {
			if lOp.LastName != rOp.LastName {
				r.operators[rOp.Number].LastName = rOp.LastName
			}

			if lOp.FirstName != rOp.FirstName {
				r.operators[rOp.Number].FirstName = rOp.FirstName
			}
		} else {
			r.operators[rOp.Number] = ops[i]
		}
	}

	for n := range r.operators {
		if _, exists := m[n]; !exists {
			delete(r.operators, n)
		}
	}

	operators := make([]*entity.Operator, 0)
	for i := range r.operators {
		operators = append(operators, r.operators[i])
	}

	return operators, nil
}

func (r *OperatorRepo) SetBusy(number string, busy bool) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if _, exists := r.operators[number]; exists {
		r.operators[number].SetBusy(busy)
	}
}

func (r *OperatorRepo) Close() {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.operators = nil
}

func NewOperatorRepo(cfg *config.OperatorRepo) *OperatorRepo {
	return &OperatorRepo{
		apiClient: &httpAPIClient{
			noVerify:   cfg.NoVerify,
			apiURL:     cfg.APIURL,
			apiTimeout: cfg.APITimeout,
		},
		operators: make(map[string]*entity.Operator),
	}
}
