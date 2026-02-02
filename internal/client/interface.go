// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import "context"

// MonitorAPI defines the interface for monitor operations.
// This interface enables mocking in unit tests.
type MonitorAPI interface {
	ListMonitors(ctx context.Context) ([]Monitor, error)
	GetMonitor(ctx context.Context, uuid string) (*Monitor, error)
	CreateMonitor(ctx context.Context, req CreateMonitorRequest) (*Monitor, error)
	UpdateMonitor(ctx context.Context, uuid string, req UpdateMonitorRequest) (*Monitor, error)
	DeleteMonitor(ctx context.Context, uuid string) error
	PauseMonitor(ctx context.Context, uuid string) (*Monitor, error)
	ResumeMonitor(ctx context.Context, uuid string) (*Monitor, error)
}

// IncidentAPI defines the interface for incident operations.
// This interface enables mocking in unit tests.
type IncidentAPI interface {
	ListIncidents(ctx context.Context) ([]Incident, error)
	GetIncident(ctx context.Context, id string) (*Incident, error)
	CreateIncident(ctx context.Context, req CreateIncidentRequest) (*Incident, error)
	UpdateIncident(ctx context.Context, id string, req UpdateIncidentRequest) (*Incident, error)
	DeleteIncident(ctx context.Context, id string) error
	AddIncidentUpdate(ctx context.Context, uuid string, req AddIncidentUpdateRequest) (*Incident, error)
	ResolveIncident(ctx context.Context, uuid string, message string) (*Incident, error)
}

// MaintenanceAPI defines the interface for maintenance operations.
// This interface enables mocking in unit tests.
type MaintenanceAPI interface {
	ListMaintenance(ctx context.Context) ([]Maintenance, error)
	GetMaintenance(ctx context.Context, id string) (*Maintenance, error)
	CreateMaintenance(ctx context.Context, req CreateMaintenanceRequest) (*Maintenance, error)
	UpdateMaintenance(ctx context.Context, id string, req UpdateMaintenanceRequest) (*Maintenance, error)
	DeleteMaintenance(ctx context.Context, id string) error
}

// ReportsAPI defines the interface for reporting operations.
// This interface enables mocking in unit tests.
type ReportsAPI interface {
	GetMonitorReport(ctx context.Context, uuid string, from, to string) (*MonitorReport, error)
	ListMonitorReports(ctx context.Context, from, to string) ([]MonitorReport, error)
}

// OutageAPI defines the interface for outage operations.
// This interface enables mocking in unit tests.
type OutageAPI interface {
	GetOutage(ctx context.Context, uuid string) (*Outage, error)
	ListOutages(ctx context.Context) ([]Outage, error)
	CreateOutage(ctx context.Context, req CreateOutageRequest) (*Outage, error)
	AcknowledgeOutage(ctx context.Context, uuid string) (*OutageAction, error)
	UnacknowledgeOutage(ctx context.Context, uuid string) (*OutageAction, error)
	ResolveOutage(ctx context.Context, uuid string) (*OutageAction, error)
	EscalateOutage(ctx context.Context, uuid string) (*OutageAction, error)
	DeleteOutage(ctx context.Context, uuid string) error
}

// HealthcheckAPI defines the interface for healthcheck operations.
// This interface enables mocking in unit tests.
type HealthcheckAPI interface {
	GetHealthcheck(ctx context.Context, uuid string) (*Healthcheck, error)
	ListHealthchecks(ctx context.Context) ([]Healthcheck, error)
	CreateHealthcheck(ctx context.Context, req CreateHealthcheckRequest) (*Healthcheck, error)
	UpdateHealthcheck(ctx context.Context, uuid string, req UpdateHealthcheckRequest) (*Healthcheck, error)
	DeleteHealthcheck(ctx context.Context, uuid string) error
	PauseHealthcheck(ctx context.Context, uuid string) (*HealthcheckAction, error)
	ResumeHealthcheck(ctx context.Context, uuid string) (*HealthcheckAction, error)
}

// StatusPageAPI defines the interface for status page operations.
// This interface enables mocking in unit tests.
type StatusPageAPI interface {
	ListStatusPages(ctx context.Context, page *int, search *string) (*StatusPagePaginatedResponse, error)
	GetStatusPage(ctx context.Context, uuid string) (*StatusPage, error)
	CreateStatusPage(ctx context.Context, req CreateStatusPageRequest) (*StatusPage, error)
	UpdateStatusPage(ctx context.Context, uuid string, req UpdateStatusPageRequest) (*StatusPage, error)
	DeleteStatusPage(ctx context.Context, uuid string) error
	ListSubscribers(ctx context.Context, uuid string, page *int, subscriberType *string) (*SubscriberPaginatedResponse, error)
	AddSubscriber(ctx context.Context, uuid string, req AddSubscriberRequest) (*StatusPageSubscriber, error)
	DeleteSubscriber(ctx context.Context, uuid string, subscriberID int) error
}

// HyperpingAPI combines all API interfaces for the Hyperping client.
// This is the primary interface used by the provider resources.
type HyperpingAPI interface {
	MonitorAPI
	IncidentAPI
	MaintenanceAPI
	ReportsAPI
	OutageAPI
	HealthcheckAPI
	StatusPageAPI
}

// Ensure Client implements HyperpingAPI at compile time.
var _ HyperpingAPI = (*Client)(nil)
