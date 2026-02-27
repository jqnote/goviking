# Context Database

## Overview

Core context database functionality with tiered loading (L0/L1/L2) for efficient context management in AI agents.

## ADDED Requirements

### Requirement: Tiered Context Loading
The system SHALL support three tiers of context (L0, L1, L2) with different load strategies.

#### Scenario: L0 Always Loaded
- **WHEN** a session is created or resumed
- **THEN** L0 (essential context) SHALL be automatically loaded

#### Scenario: L1 Loaded on Demand
- **WHEN** L1 context is explicitly requested by the agent
- **THEN** L1 context SHALL be loaded from storage

#### Scenario: L2 Loaded on Demand
- **WHEN** L2 context is explicitly requested
- **THEN** L2 context SHALL be loaded from storage

### Requirement: Context Window Management
The system SHALL manage context window limits by prioritizing and compressing context as needed.

#### Scenario: Context Exceeds Window
- **WHEN** total context exceeds the model's window limit
- **THEN** the system SHALL prioritize higher-tier context over lower-tier

#### Scenario: Automatic Compression
- **WHEN** context approaches window limit
- **THEN** the system SHALL offer to compress lower-priority context

### Requirement: Context Builder
The system SHALL provide a context builder for constructing context from multiple sources.

#### Scenario: Build Context from Multiple Sources
- **WHEN** context is requested from memories, resources, and skills
- **THEN** the system SHALL merge them into a unified context

### Requirement: Context Persistence
The system SHALL persist context data to storage for session recovery.

#### Scenario: Save Context
- **WHEN** session state changes
- **THEN** the system SHALL persist context to storage

#### Scenario: Restore Context
- **WHEN** session is resumed
- **THEN** the system SHALL restore persisted context from storage
