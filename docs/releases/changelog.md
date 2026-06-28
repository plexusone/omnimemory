# Changelog

All notable changes to OmniMemory are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

No unreleased changes.

## [0.1.0] - 2024-06-27

Initial release.

### Added

- Core types and provider interface
- Multi-provider client with fallback support
- Provider registry with priority-based selection
- Conformance test suite (`core/providertest`)
- In-memory provider for testing
- PostgreSQL provider with pgvector support
- KVS provider wrapping omnistorage-core
- Provider stubs for Mem0, Graphiti, Twilio
- SessionID field in Context and Memory
- Embedder interface with OmniLLM integration
- Memory scopes (user, agent, tenant, team, session, domain)
- Memory types (observation, fact, preference, summary, trait, relationship)
- Ent schema for PostgreSQL storage
- MkDocs documentation site

### Fixed

- Context parameter usage in client fallback
- Subject isolation in memory provider

[0.1.0]: https://github.com/plexusone/omnimemory/releases/tag/v0.1.0

---

For external provider changelogs:

- [omni-twilio/omnimemory](https://github.com/plexusone/omni-twilio/blob/main/CHANGELOG.md)
