# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.1.0] - 2026-09-30

### Added

- Initial Go project structure.
- User registration with unique usernames and bcrypt password hashing.
- User authentication with in-memory sessions.
- Public posts feed.
- Post creation for authenticated users.
- Post publication timestamps, preserving unknown dates for older posts.
- Russian and English localization with a persistent language preference.
- Localized date and time formatting with explicit UTC display.
- PostgreSQL database support with automatic table setup.
- Docker Compose configuration for the development database.
- Project README, concise agent instructions, and a separate project context guide
  for AI agents.
