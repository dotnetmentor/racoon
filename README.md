# racoon - secrets are my thing

## Commands

See `racoon help` or ` racoon --help` for all available commands

## Sources

- AWS Systems Manager : Parameter Store

## Outputs

- dotenv
- Terraform tfvars

## Roadmap

- Seeding of secrets not already in the store
- Release pipeline
- Export outputs to stdout (no logging allowed)
- Reading a single secret
- Context support (dev / production / cicd / localdev etc)
- Key format for Parameter Store
- Generators for providing generated values when seeding a secret
- Listing secrets in a given context
- Deleting a secret from the store
- Conditional sync for faster exports (export based on hash sum for context)
- Shell (bash/zsh/sh) output format
- Certificate output format
- Kubernetes secret output format
- Naming conventions for outputs
- Command for local cleanup of generated files
- Store provider for AWS Secrets Manager : Secrets
