# racoon - secrets are my thing

## Commands

See `racoon help` or ` racoon --help` for all available commands

## Sources

- AWS Systems Manager : Parameter Store

## Outputs

- dotenv
- json
- tfvars (Terraform)

## Examples

### Commands

```bash
racoon create                                   # ensures secrets missing in the remote store are created by prompting the user for input
racoon read MongodbConnection                   # reads a single secret from the remote store and writes it's value to stdout
racoon export                                   # exports all secrets using the outputs defines in the manifest file
racoon export --output direnv                   # exports all secrets using the direnv output defined in the manifest file
racoon export --output direnv --path dot.env    # exports all secrets using the direnv output to the specified path
racoon export -o direnv -p -                    # exports all secrets using the direnv output, writing the result to stdout
racoon export -o direnv --include Secret1       # export Secret1 using the direnv output
racoon export -o direnv --exclude Secret1       # export all secrets but Secret1 using the direnv output
```

### secrets.y\*ml

```yaml
stores:
  awsParameterStore:
    kmsKey: alias/parameter_store_key
    keyFormat: "/{Context}/{Key}"
secrets:
  - name: MongodbConnection
    description: MongoDB Connection string
    valueFrom:
      awsParameterStore:
        key: /fixed/key/for/mongodb/connection
  - name: TwilioAccountSid
    description: Twilio Account ID
    valueFrom:
      awsParameterStore: {}
  - name: TwilioAuthToken
    description: Twilio Auth Token
    valueFrom:
      awsParameterStore: {}
  - name: TwilioServiceId
    description: Twilio Auth Token
    valueFrom:
      awsParameterStore: {}
  - name: DefaultSender
    description: The default sender email address
    default: noreply@mydomain.com
outputs:
  - type: dotenv
    path: output/.env
  - type: tfvars
    path: output/secrets.tfvars
    exclude:
      - MongodbConnection
```

## Roadmap

- [x] Exporting of secrets to multiple outputs (dotenv, tfvars)
- [x] Seeding of secrets not already in the store
- [ ] Release pipeline
- [x] Export outputs to stdout (no logging allowed)
- [x] Command for reading a single secrets value
- [ ] Tagging of external resources
- [x] Context support (dev / production / cicd / localdev etc)
- [x] Key format for Parameter Store
- [x] Remapping support for outputs (PaymetApiKey -> Payment\_\_ApiKey)
- [ ] Generators for providing generated values when seeding a secret
- [ ] Listing secrets in a given context
- [ ] Deleting a secret from the store
- [x] Json output format
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
- [x] Cleaner handling of errors (less panic, more logging and exit codes)
- [x] Ability to select secrets for export using flags (racoon export --include||--exclude Secret1)
- [x] Ability to select secrets for export using output config (include:[] exclude:[])
- [ ] Conditional sync for faster exports (export based on hash sum for context)
- [ ] Dump command for creating a local cache that is used until cache is no longer available (no calls to the store when cache is available)

## Development

```sh
go get
go run . -- <args>
```
