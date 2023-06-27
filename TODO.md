# TODO

- [x] Should we only be using the layers concept and not allow properties at the manifest root (would enable naming the "base" layer to something else)? No
- [x] Renamed config.sources (in manifest) to config.defaults? No.
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

- [ ] (in-progress) Re-implement create command (writable sources)
- [ ] (in-progress) General refactor of "source.go" etc, possibly new package
- [ ] (in-progress) Resolve TODO's in code
- [ ] Rename "layer" flag and alias to "parameter" and "p"

- [ ] Get test cases working again
- [ ] Allow writing to writable sources defined by formatting config

- [ ] Caching for sources during a single run
- [ ] Allow layers to be defined in separate files
- [ ] Validation options: Value type (Int, String, Bool etc)
- [ ] Validation options: Value match Regexp (.\*)
- [ ] Use config.sources as a way to enable the use of a source (if not specified, then it's not enabled)
- [ ] How do we express that an empty value is OK for a source? Do we? (Using validation field on property?)
- [ ] Allow generating values to help with seeding the store
