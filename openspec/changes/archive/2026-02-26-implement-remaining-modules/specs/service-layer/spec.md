# Service Layer

## ADDED Requirements

### Requirement: Business Logic
The system SHALL encapsulate business logic in service layer.

#### Scenario: Context Service
- **WHEN** context operations are called
- **THEN** business logic SHALL be applied

### Requirement: Validation
The system SHALL validate inputs before processing.

#### Scenario: Validate Context
- **WHEN** invalid context is submitted
- **THEN** error SHALL be returned
