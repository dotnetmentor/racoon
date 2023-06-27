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
- [x] Basic layer regex matching
- [x] Re-implement create command (renamed to write)
- [x] General refactor
- [x] Resolve important TODO's in code
- [x] Rename "layer" flag and alias to "parameter" and "p"
- [x] Rename deny* to allow* (allowImplicit: true), default must change
- [x] Remove "current value" from prompt for new values (write command)
- [x] Bug, do not ask about preview twice (when using write command)
- [x] Bug, when setting new value, never allow log of sensitive value

- [ ] (in-progress) Initial round of real world testing
- [ ] Get test cases working again
- [ ] Allow writing to writable sources defined by formatting config

- [ ] Warn about trying to "redefine" rules of a property
- [ ] Warn about finding value for property that does not allow implicit overrides
- [ ] Warn about defining explicit property in layer when explicity overrides is not allowed by parent property rules
- [ ] Validation options: Value type (Int, String, Bool etc)
- [ ] Validation options: Value match Regexp (.\*)
- [ ] Validate that ImplicitSources list is unique
- [ ] Parameter validation (minLength: 3, maxLength: 16 etc)
- [ ] New dot-format for propety names to enable structured output in json etc (name: Translation.TravelwebUrl -> {"translation": {"travelWebUrl": "..."}})?
- [ ] Caching for sources during a single run
- [ ] Allow layers to be defined in separate files
- [ ] Use config.sources as a way to enable the use of a source (if not specified, then it's not enabled)?
- [ ] How do we express that an empty value is OK for a source? Do we? (Using validation field on property?)
- [ ] Allow generating values to help with seeding the store
- [ ] Allow enforcing senitive values can't be written to unsafe store
- [ ] Option to sort output keys (dotenv)
- [ ] Make formatting rules a list to get predictable order
- [ ] Add support for regexReplace (regexReplace: "/demo-\*//")
