package seeders

import (
	"context"
	"database/sql"
	"fmt"
)

type Seeder interface {
	Name() string
	Run(ctx context.Context, db *sql.DB) error
}

type Registry struct {
	seeders map[string]Seeder
	order   []string
}

func NewRegistry() *Registry {
	r := &Registry{
		seeders: make(map[string]Seeder),
		order:   make([]string, 0),
	}
	r.register(&RoleSeeder{})
	r.register(&UserSeeder{})
	return r
}

func (r *Registry) register(s Seeder) {
	r.seeders[s.Name()] = s
	r.order = append(r.order, s.Name())
}

func (r *Registry) RunAll(ctx context.Context, db *sql.DB) error {
	for _, name := range r.order {
		if err := r.seeders[name].Run(ctx, db); err != nil {
			return fmt.Errorf("seeder %s: %w", name, err)
		}
	}
	return nil
}

func (r *Registry) Run(ctx context.Context, db *sql.DB, name string) error {
	s, ok := r.seeders[name]
	if !ok {
		return fmt.Errorf("seeder %s tidak ditemukan", name)
	}
	return s.Run(ctx, db)
}
