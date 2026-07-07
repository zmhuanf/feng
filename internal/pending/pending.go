package pending

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Result struct {
	Success bool
	Data    any
}

type Request struct {
	ID       string
	Callback any
	ch       chan Result
	once     sync.Once
}

type Store struct {
	timeout time.Duration
	items   map[string]*Request
	lock    sync.RWMutex
}

func New(timeout time.Duration) *Store {
	return &Store{
		timeout: timeout,
		items:   make(map[string]*Request),
	}
}

func (s *Store) Add(callback any) *Request {
	req := &Request{
		ID:       uuid.New().String(),
		Callback: callback,
		ch:       make(chan Result, 1),
	}
	s.lock.Lock()
	s.items[req.ID] = req
	s.lock.Unlock()
	return req
}

func (s *Store) Get(id string) (*Request, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	req, ok := s.items[id]
	return req, ok
}

func (s *Store) Delete(id string) {
	s.lock.Lock()
	req, ok := s.items[id]
	if ok {
		delete(s.items, id)
	}
	s.lock.Unlock()
	if ok {
		req.close()
	}
}

func (s *Store) Resolve(id string, result Result) bool {
	s.lock.Lock()
	req, ok := s.items[id]
	if ok {
		delete(s.items, id)
	}
	s.lock.Unlock()
	if !ok {
		return false
	}
	req.ch <- result
	req.close()
	return true
}

func (s *Store) Wait(ctx context.Context, req *Request) error {
	timer := time.NewTimer(s.timeout)
	defer timer.Stop()

	select {
	case result, ok := <-req.ch:
		if !ok {
			return errors.New("request closed")
		}
		if result.Success {
			return nil
		}
		return fmt.Errorf("%v", result.Data)
	case <-ctx.Done():
		s.Delete(req.ID)
		return ctx.Err()
	case <-timer.C:
		s.Delete(req.ID)
		return errors.New("request timeout")
	}
}

func (s *Store) AutoDelete(req *Request) {
	timer := time.NewTimer(s.timeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		s.Delete(req.ID)
	case <-req.ch:
		req.close()
	}
}

func (s *Store) Close() {
	s.lock.Lock()
	items := s.items
	s.items = make(map[string]*Request)
	s.lock.Unlock()

	for _, req := range items {
		req.close()
	}
}

func (r *Request) close() {
	r.once.Do(func() { close(r.ch) })
}
