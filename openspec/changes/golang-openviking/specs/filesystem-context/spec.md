# Filesystem Context (AGFS)

## Overview

Filesystem-like context management (AGFS - Agent Graph File System) for unified memory, resources, and skills organization.

## ADDED Requirements

### Requirement: Directory Structure
The system SHALL organize context using a hierarchical directory structure.

#### Scenario: Create Directory
- **WHEN** user creates a directory in the context filesystem
- **THEN** the directory SHALL be persisted and visible in the tree

#### Scenario: List Directory
- **WHEN** user lists contents of a directory
- **THEN** the system SHALL return all direct children (files and subdirectories)

### Requirement: Context as Files
The system SHALL represent memories, resources, and skills as files in the filesystem.

#### Scenario: Create Memory File
- **WHEN** user creates a memory entry
- **THEN** it SHALL be represented as a file in the memories directory

#### Scenario: Create Resource File
- **WHEN** user adds a resource
- **THEN** it SHALL be represented as a file in the resources directory

#### Scenario: Create Skill File
- **WHEN** user adds a skill definition
- **THEN** it SHALL be represented as a file in the skills directory

### Requirement: File Operations
The system SHALL support basic file operations (read, write, delete, move).

#### Scenario: Read File
- **WHEN** user reads a file by path
- **THEN** the system SHALL return the file content

#### Scenario: Write File
- **WHEN** user writes to a file path
- **THEN** the content SHALL be persisted

#### Scenario: Delete File
- **WHEN** user deletes a file
- **THEN** the file SHALL be removed from the filesystem

### Requirement: Tree Structure
The system SHALL maintain a tree structure for navigation and retrieval.

#### Scenario: Get Tree
- **WHEN** user requests the full directory tree
- **THEN** the system SHALL return a nested structure representing all context
