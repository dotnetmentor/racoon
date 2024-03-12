# racoon

Layered Configuration as Code

## Commands

See `racoon --help` for all available commands

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
racoon read MongodbConnection                   # reads a single value and writes it's value to stdout
racoon export                                   # exports all values using the outputs defines in the manifest file
racoon export --output direnv                   # exports all values using the direnv output defined in the manifest file
racoon export --output direnv --path dot.env    # exports all values using the direnv output to the specified path
racoon export -o direnv -p -                    # exports all values using the direnv output, writing the result to stdout
racoon export -o direnv --include Secret1       # export Secret1 using the direnv output
racoon export -o direnv --exclude Secret1       # export all values but Secret1 using the direnv output
```

### racoon.y\*ml

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
    config:
      quote: false
  - type: tfvars
    path: output/secrets.tfvars
    exclude:
      - MongodbConnection
```

## Roadmap

[Check out the roadmap](https://github.com/orgs/dotnetmentor/projects/1)
