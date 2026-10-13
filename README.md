# xcl-plugin-test

A test-only [xcl](https://github.com/jumppad-labs/xcl) plugin. It is published
as signed GitHub releases so that xcl and its language bindings can test
against a real, externally released plugin: the GitHub registry's download,
checksum and signature checks, a plugin started from a path or a directory,
and the resource lifecycle, sensitive values and computed values through a
plugin program. It needs no outside service, so it runs the same on every
platform.

It is not an example to copy. To write a plugin, start from
[xcl-plugin-template](https://github.com/jumppad-labs/xcl-plugin-template); for
a worked plugin for a real backend, see
[xcl-plugin-docker](https://github.com/jumppad-labs/xcl-plugin-docker). This
plugin keeps the template's layout, so it is built, tested and released the
same way.

## The block type

The plugin provides one block type, `test "item"`, which does nothing but
record each operation xcl asks of it:

```hcl
test "item" "first" {
  log    = "./ops.jsonl"
  value  = "one"
  secret = "hunter2"
}

test "item" "second" {
  log   = "./ops.jsonl"
  value = "two"
  input = test.item.first.output
}
```

| Field    | Kind                        | Meaning |
|----------|-----------------------------|---------|
| `log`    | string, required            | The JSON-lines file every operation on the item appends its record to. |
| `value`  | string, optional            | A plain setting, updated in place. |
| `input`  | string, optional            | Usually another item's `output`, which makes this item depend on that one. Updated in place. |
| `secret` | sensitive string, optional  | Redacted by xcl in events and lookups, masked in state. Updated in place. |
| `output` | computed string             | `<name>-<value>`, set by create and update. |

Every setting is updated in place: an item is never replaced. `examples/basic`
holds the two items above.

## The operation log

Create, read, update and destroy each append one line to the item's `log`
before doing their work, which is only to compute `output`:

```json
{"operation":"create","address":"test.item.first"}
```

Lines are appended in the order the calls start, and the log outlives the
item, so a test can read back a destroy. xcl runs items that do not depend on
each other concurrently, so only compare the order of linked items: above,
`first` is created before `second` and destroyed after it. Reads happen as
often as xcl needs them; count creates, updates and destroys, not reads.

## Layout

- `entities` — the `Item` block type, with no behaviour, so an application that
  reads applied items imports it alone.
- `client/oplog` — the operation log, behind the `OperationLog` interface, with
  its Mockery double in `client/oplog/mocks`.
- `providers` — the item provider, which records through the operation log.
- `plugin.go` — `testplugin.Plugin`, which registers the type in-process.
- `cmd/xcl-plugin-test` — the plugin program, which serves the same plugin over
  gRPC.
- `e2e` — applies, plans, updates and destroys `examples/basic` in-process and
  as a separate program, and reads the log back.

## Developing

```sh
make build      # build every package and build/xcl-plugin-test
make test       # unit and end-to-end tests
make lint       # go vet, gofmt and staticcheck
make generate   # regenerate the Mockery double
make dist VERSION=v0.1.0   # release archives and checksums in dist/
```

## Releasing

Releases are cut by hand. On `main`, with CI green:

```sh
make release VERSION=v0.1.0
```

This tags the commit and pushes the tag. The release workflow then runs the
tests, builds an archive per platform (`xcl-plugin-test_<version>_<os>_<arch>.tar.gz`,
`.zip` on Windows), writes `xcl-plugin-test_<version>_checksums.txt`, signs it
with the jumppad-labs release key into `xcl-plugin-test_<version>_checksums.txt.sig`,
and publishes the release. It needs the `GPG_PRIVATE_KEY` and `GPG_PASSPHRASE`
repository secrets.

The xcl bindings pin the version they install; bumping it is a change in each
binding, for xcl-python `tests/plugins.toml`.

## License

Apache 2.0, see [LICENSE](LICENSE).
