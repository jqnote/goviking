# CLI Tool

## Overview

Command-line interface for context database management.

## ADDED Requirements

### Requirement: CLI Commands
The system SHALL provide a command-line interface for all core operations.

#### Scenario: View Help
- **WHEN** user runs `goviking --help`
- **THEN** all available commands SHALL be displayed

### Requirement: Context Commands
The system SHALL provide commands for context management.

#### Scenario: List Context
- **WHEN** user runs `goviking context list`
- **THEN** all context entries SHALL be displayed

#### Scenario: Show Context
- **WHEN** user runs `goviking context show <id>`
- **THEN** the context entry SHALL be displayed

#### Scenario: Create Context
- **WHEN** user runs `goviking context create <path>`
- **THEN** a new context entry SHALL be created

### Requirement: Session Commands
The system SHALL provide commands for session management.

#### Scenario: List Sessions
- **WHEN** user runs `goviking session list`
- **THEN** all sessions SHALL be displayed

#### Scenario: Show Session
- **WHEN** user runs `goviking session show <id>`
- **THEN** the session details SHALL be displayed

#### Scenario: Resume Session
- **WHEN** user runs `goviking session resume <id>`
- **THEN** the session SHALL be restored

### Requirement: Filesystem Commands
The system SHALL provide commands for filesystem operations.

#### Scenario: Tree View
- **WHEN** user runs `goviking fs tree`
- **THEN** the directory tree SHALL be displayed

#### Scenario: List Files
- **WHEN** user runs `goviking fs ls <path>`
- **THEN** files at the path SHALL be listed

### Requirement: Retrieval Commands
The system SHALL provide commands for retrieval operations.

#### Scenario: Search
- **WHEN** user runs `goviking search <query>`
- **THEN** search results SHALL be displayed

### Requirement: Config Commands
The system SHALL provide commands for configuration.

#### Scenario: Show Config
- **WHEN** user runs `goviking config show`
- **THEN** current configuration SHALL be displayed

#### Scenario: Init Config
- **WHEN** user runs `goviking config init`
- **THEN** a default configuration file SHALL be created
