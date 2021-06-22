// Code generated by golangee/eearc; DO NOT EDIT.
//
// Copyright 2021 Torben Schinke
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package supportietyserver

import (
	context "context"
	fmt "fmt"
	core "github.com/golangee/architecture/testdata/workspace/server/internal/tickets/core"
	usecase "github.com/golangee/architecture/testdata/workspace/server/internal/tickets/usecase"
)

// Application embeds the defaultApplication to provide the default application behavior.
// It also provides the inversion of control injection mechanism for all bounded contexts.
type Application struct {
	defaultApplication
}

func NewApplication(ctx context.Context) (*Application, error) {
	a := &Application{}
	a.defaultApplication.self = a
	if err := a.init(ctx); err != nil {
		return nil, fmt.Errorf("cannot init application: %w", err)
	}

	return a, nil
}

// defaultApplication aggregates all contained bounded contexts and starts their driver adapters.
type defaultApplication struct {
	// self provides a pointer to the actual Application instance to provide
	// one level of vtable calling indirection for simple method 'overriding'.
	self *Application

	ticketsUsecaseTickets *usecase.Tickets
}

func (_ defaultApplication) init(ctx context.Context) error {
	return nil
}

func (_ defaultApplication) Run(ctx context.Context) error {
	return nil
}

func (_ Application) init(ctx context.Context) error {
	fmt.Println("hey concrete init")
	return nil
}

func (a *Application)Run(ctx context.Context)error{
	fmt.Println("hey concrete run")
	return nil
}

func (d *defaultApplication) getTicketsUsecaseTickets(myCfg usecase.MyConfig, tickets core.Tickets) (*usecase.Tickets, error) {
	if d.ticketsUsecaseTickets != nil {
		return d.ticketsUsecaseTickets, nil
	}

	s, err := usecase.NewTickets(myCfg, tickets)
	if err != nil {
		return nil, fmt.Errorf("cannot create service 'Tickets': %w", err)
	}

	d.ticketsUsecaseTickets = s

	return s, nil
}
