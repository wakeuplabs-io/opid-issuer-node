# Changelog from v2 to v3
This is a resume of PrivadoID's changelog
- [v3.0.0](https://github.com/0xPolygonID/issuer-node/releases/tag/v3.0.0)
- [v3.0.1](https://github.com/0xPolygonID/issuer-node/releases/tag/v3.0.1)

# 🚀 Issuer Node v3 Highlights

## Unified API (v2)
- Core API and UI API have been consolidated into a single, unified API.
- Simplifies development and enhances the developer experience.
- Retains all functionalities from API v1.
- Introduces improved credential status checks.
- All endpoints (authentication, schema, link management, connections, etc.) now fall under this unified API.

## Credential Issuance Improvements
- Streamlined flow: Users can now authenticate and receive credentials in a single step using universal links.
- QR code generation updated for better integration with universal links.
- Credential endpoints renamed from "claims" to **"credentials"**.
- Added support for paginated credential retrieval and credential deletion.

## Identity & Key Management
- Display names added to identities with uniqueness constraints.
- Support for multiple identities per issuer across multiple blockchain networks.
- Identity details now include key types and display names.
- Vault setup is now optional.
- AWS Key Management Service (KMS) support introduced.

## UI Enhancements
- Refreshed browser UI with support for:
  - Creating multiple identities.
  - Issuing credentials via universal links.
- Improved rendering (e.g., fixed ellipsis bug in Safari).

# 🧰 Developer Experience & Maintenance

## API Documentation & Validation
- More detailed endpoint descriptions and example responses.
- Improved error handling (e.g., proper 400 errors for invalid subject IDs).

## Codebase & Dev Tools
- Updated dependencies, Makefile, and folder structure.
- Removed outdated references (e.g., to Polygon ID).

## Docker & Setup
- Fixed `docker-compose-full` issues for smoother local development.

# 🔧 Miscellaneous Fixes & Additions

## Swagger & Link Generation
- UI improvements and consistency fixes in universal/deeplink responses.

## Supported Networks Endpoint
- Added accepted RHS (Revocation Handling Service) modes.
- Renamed `rhsMode` to `credentialStatus` for better clarity.

