# Client SDK

## ADDED Requirements

### Requirement: Sync Client
The system SHALL provide a synchronous client.

#### Scenario: Create Client
- **WHEN** user creates a client
- **THEN** client SHALL be ready to use

### Requirement: Async Client
The system SHALL provide an asynchronous client.

#### Scenario: Create Async Client
- **WHEN** user creates an async client
- **THEN** client SHALL support async operations

### Requirement: Context Operations
The system SHALL support context CRUD via client.

#### Scenario: Create Context via Client
- **WHEN** user calls CreateContext
- **THEN** context SHALL be created

### Requirement: Session Operations
The system SHALL support session operations via client.

#### Scenario: Create Session via Client
- **WHEN** user calls CreateSession
- **THEN** session SHALL be created
