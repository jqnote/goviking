# Server API

## ADDED Requirements

### Requirement: REST Endpoints
The system SHALL provide REST API endpoints.

#### Scenario: Start Server
- **WHEN** server is started
- **THEN** REST endpoints SHALL be available

### Requirement: Context Endpoints
The system SHALL provide CRUD endpoints for contexts.

#### Scenario: POST /contexts
- **WHEN** POST request to /contexts
- **THEN** context SHALL be created

#### Scenario: GET /contexts
- **WHEN** GET request to /contexts
- **THEN** list of contexts SHALL be returned

### Requirement: Session Endpoints
The system SHALL provide session management endpoints.

#### Scenario: POST /sessions
- **WHEN** POST request to /sessions
- **THEN** session SHALL be created
