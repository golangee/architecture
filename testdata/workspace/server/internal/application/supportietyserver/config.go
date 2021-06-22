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
	usecase "github.com/golangee/architecture/testdata/workspace/server/internal/tickets/usecase"
)

// Configuration contains all aggregated configurations for the entire application and all contained bounded contexts.
type Configuration struct {
	TicketsConfig TicketsConfig
}

// TicketsConfig contains all aggregated configurations for the entire bounded context 'tickets'.
type TicketsConfig struct {
	TicketsUsecaseConfig TicketsUsecaseConfig
}

// TicketsUsecaseConfig contains all configurations for the 'usecase' layer of 'tickets'.
type TicketsUsecaseConfig struct {
	MyConfig usecase.MyConfig
}
