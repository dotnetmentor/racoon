# racoon - secrets are my thing

## Commands

See `racoon help` or ` racoon --help` for all available commands

## Sources

- AWS Systems Manager : Parameter Store

## Outputs

- dotenv
- Terraform tfvars

## Examples

### Commands

```bash
racoon create                                   # ensures secrets missing in the remote store are created by prompting the user for input
racoon read MongodbConnection                   # reads a single secret from the remote store and writes it's value to stdout
racoon export                                   # exports all secrets using the outputs defines in the manifest file
racoon export --output direnv                   # exports all secrets using the direnv output defined in the manifest file
racoon export --output direnv --path dot.env    # exports all secrets using the direnv output to the specified path
racoon export -o direnv -p -                    # exports all secrets using the direnv output, writing the result to stdout
```

### secrets.y\*ml

```yaml
stores:
  awsParameterStore:
    kmsKey: alias/parameter_store_key
secrets:
  - name: MongodbConnection
    description: MongoDB Connection string
    valueFrom:
      awsParameterStore:
        key: /localdev/mongodb/connection
  - name: TwilioAccountSid
    description: Twilio Account ID
    valueFrom:
      awsParameterStore:
        key: /localdev/twilio/account_sid
  - name: TwilioAuthToken
    description: Twilio Auth Token
    valueFrom:
      awsParameterStore:
        key: /localdev/twilio/auth_token
  - name: TwilioServiceId
    description: Twilio Auth Token
    valueFrom:
      awsParameterStore:
        key: /localdev/twilio/service_id
  - name: DefaultSender
    description: The default sender email address
    default: noreply@mydomain.com
outputs:
  - type: dotenv
    path: output/.env
```

## Roadmap

- [x] Exporting of secrets to multiple outputs (dotenv, tfvars)
- [x] Seeding of secrets not already in the store
- [ ] Release pipeline
- [x] Export outputs to stdout (no logging allowed)
- [x] Command for reading a single secrets value
- [ ] Tagging of external resources
- [ ] Context support (dev / production / cicd / localdev etc)
- [ ] Key format for Parameter Store
- [ ] Generators for providing generated values when seeding a secret
- [ ] Listing secrets in a given context
- [ ] Deleting a secret from the store
- [ ] Json output format
- [ ] Shell (bash/zsh/sh) output format
- [ ] Certificate output format
- [ ] Kubernetes secret output format
- [ ] Naming conventions for outputs
- [ ] Command for local cleanup of generated files
- [ ] Store provider for AWS Secrets Manager : Secrets
- [ ] Store provider for Azure Key Vault : Secrets
- [ ] Flag for specifying other filenames for secrets.y\*ml
- [ ] Readonly secrets (used for consuming secret managed by external system)
- [ ] Move command for moving secrets in the store
- [ ] Init command for creating the manifest file
- [ ] Cleaner handling of errors (less panic, more logging and exit codes)
- [ ] Ability to select secrets for export (racoon export -s Secret1 -s Secret2)
- [ ] Conditional sync for faster exports (export based on hash sum for context)

## Development

```sh
go get
go run . -- <args>
```
