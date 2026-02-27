# Config Management

## ADDED Requirements

### Requirement: YAML Configuration
The system SHALL load configuration from YAML files.

#### Scenario: Load YAML Config
- **WHEN** config file exists at default location
- **THEN** configuration SHALL be loaded from YAML

#### Scenario: Custom Config Path
- **WHEN** custom config path is provided
- **THEN** configuration SHALL be loaded from that path

### Requirement: Environment Variables
The system SHALL override config values with environment variables.

#### Scenario: Env Override
- **WHEN** environment variable matches config key
- **THEN** env value SHALL override config file value

### Requirement: Default Values
The system SHALL provide sensible default values.

#### Scenario: Use Defaults
- **WHEN** config file is missing
- **THEN** default values SHALL be used
