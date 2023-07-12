# TODO

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

- [ ] (in-progress) Initial round of real world testing

- [ ] Option to sort output keys (dotenv)

- [ ] More tests on multiple levels and components (e2e, unit etc)
- [ ] Feature: New dot-format for propety names to enable structured output in json etc (name: Translation.TravelwebUrl -> {"translation": {"travelWebUrl": "..."}})?
- [ ] Feature: Validation options, Value type (Int, String, Bool etc)
- [ ] Feature: Validation options, Value match Regexp (.\*)
- [ ] Feature: Validation options, String values - MinLength: 3, MaxLength: 16 etc
- [ ] Feature: Caching for sources during a single run
- [ ] Feature: Allow layers to be defined in separate files
- [ ] Feature: Use config.sources as a way to enable the use of a source (if not specified, then it's not enabled)?
- [ ] Feature: Allow generating values to help with seeding the store
- [ ] Feature: Allow enforcing senitive values can't be written to "unsafe" store
- [ ] Feature: run export with --watch to have racoon running in the background, watching for changes
