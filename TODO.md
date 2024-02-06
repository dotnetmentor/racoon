# TODO

## Completed from old roadmap

- [x] Exporting of secrets to multiple outputs (dotenv, tfvars)
- [x] Seeding of secrets not already in the store
- [x] Export outputs to stdout (no logging allowed)
- [x] Command for reading a single secrets value
- [x] Context support (dev / production / cicd / localdev etc)
- [x] Key format for Parameter Store
- [x] Remapping support for outputs (PaymetApiKey -> Payment\_\_ApiKey)
- [x] Json output format
- [x] Flag for specifying other filenames for racoon.y\*ml
- [x] Cleaner handling of errors (less panic, more logging and exit codes)
- [x] Ability to select secrets for export using flags (racoon export --include||--exclude Secret1)
- [x] Ability to select secrets for export using output config (include:[] exclude:[])
- [x] Configuration of outputs (example: dotenv without doublequotes)
- [x] Update description on existing secrets

## Completed

- [x] Introduced properties concept (replaces secrets)
- [x] Introduced the concept of sensitive values
- [x] Introduced layers concept
- [x] Allow changing the log-level using flag
- [x] Allow defining a property in a layer
- [x] Enforce sensitive value from a source (awsParameterStore etc)
- [x] Allow changing a property to "sensitive" in upper layer
- [x] Validate parameters as specified by manifest file (required or just defined)
- [x] Verify explicit overrides works
- [x] Allow formatting of a value by replacing keys as specified by the property
- [x] Allow implicit override from sources specified in a layer where property is defined in parent layer
- [x] Config: Deny implicit overrides from any layer above
- [x] Config: Deny explicit overrides from any layer above
- [x] How do we express that an empty value is OK for a property? (use default: ""?)
- [x] Allow output filter (all, sensitive, cleartext)
- [x] Re-implement read command
- [x] Basic layer regexp matching
- [x] Re-implement create command (renamed to write)
- [x] General refactor
- [x] Resolve important TODO's in code
- [x] Rename "layer" flag and alias to "parameter" and "p"
- [x] Rename deny* to allow* (allowImplicit: true), default must change
- [x] Remove "current value" from prompt for new values (write command)
- [x] Bug, do not ask about preview twice (when using write command)
- [x] Bug, when setting new value, never allow log of sensitive value
- [x] Parameter validation (regexp)
- [x] Basic test cases
- [x] Add support for regexpReplace (regexpReplace: "/demo-\*//")
- [x] Allow writing fortatter sources as defined by formatting config
- [x] Allow specifying <not-found> value to be treated as error by source (only AwsParameterStore supported at the moment)
- [x] Improved layer matching and error handling
- [x] Warn about trying to "override" rules of a property
- [x] Warn about trying to "override" description of a property
- [x] Warn about defining explicit property in layer when explicity overrides is not allowed by parent property rules
- [x] Validate that ImplicitSources list is unique
- [x] Option to sort output keys (dotenv)
- [x] UI command with web interface that allows comparing results between different exports
- [x] Support multiple paths with formatting for outputs (using --path= should override all paths specified in manifest)
- [x] Feature: Dot-based property name format for grouping and to enable structured output in json etc (name: Translation.TravelwebUrl -> {"Translation": {"TravelWebUrl": "..."}})?
- [x] Allow extending a base config, referencing a yaml file to serve as the base
- [x] Tagging of AWS SSM parameters using default tags (owner + version) and labels
- [x] Fix: Allow base config to define layers (currently, base layers are replaced by layers in defined by referencing config)
- [x] Fix: If base config has layers but referencing config does not, validation error for "duplicate layer" is triggered.
- [x] Added config show command to display the final configuration (after merge with base configs)
- [x] Basic implementation of a backend for configurations (with support for AWS S3 and KMS)
- [x] Basic UI for view backend configurations
- [x] Allow enabling and disabling backend usage using RACOON_BACKEND_ENABLED=true/false
- [x] Fix: Read command must match argument to a single property, no property match
- [x] Fix: write command needs log message when no properties have a writable source
- [x] Fix: Remove support for optional properties, define them in lower layer instead??? should be an error
- [x] Fix: Bad logging (racoon WARN[0000] dotenv file local.env was not found ... racoon DEBU[0000] dotenv file local.env loaded)
- [x] Feature: Allow prefix for dotenv output (could be used to do "export FOO=bar" or "MYSVC_FOO=bar")
- [x] Feature: Allow {name} to be replaced with the manifest name
- [x] Feature: Added logging of provided parameters during matching
- [x] Feature: Optional formatters where replacement can be enforced by defining rules
- [x] Feature: "config init" command for generating a "started" config

## In progress

- [ ] Initial round of real world testing

## Next

- [ ] What's the tagline for the project, update readme, repository and cli help

## Proposals

NOTE! These have yet to make it onto the project board

- [ ] Update the readme and move remaining todo's to roadmap
- [ ] More and better tests on multiple levels and components (e2e, unit etc)
- [ ] Feature: Basic support for Int, String and Boolean values
- [ ] Feature: Validation options, Value type (Int, String, Bool etc)
- [ ] Feature: Validation options, Value match Regexp (.\*)
- [ ] Feature: Validation options, String values - MinLength: 3, MaxLength: 16 etc
- [ ] Feature: Allow layers to be defined in separate files
- [ ] Feature: Use config.sources as a way to enable the use of a source (if not specified, then it's not enabled)?
- [ ] Feature: Add output type "merge", that combines aliased outputs
- [ ] Feature: Conditional outputs, based on same matching method as layers
- [ ] Feature: Command for listing properties
- [ ] Feature: Deleting a value from a writable source (useful for cleanup)
- [ ] Feature: Moving a value from one source to another
- [ ] Feature: Copying a value from one source to another
- [ ] Feature: Certificate output format
- [ ] Feature: Kubernetes secret output format
- [ ] Feature: Kubernetes configmap output format
- [ ] Feature: "Naming" conventions for outputs
- [ ] Feature: New writable source, AWS Secrets Manager
- [ ] Feature: New writable source, Azure Key Vault
- [ ] Feature: Readonly properties (used for consuming values managed by external system)
