# SQLC Zombiezen driver plugin

## How to use

1. Install the plugin:
   ```shell
   go install github.com/delaneyj/toolbelt/sqlc-gen-zombiezen@latest
   ```
2. Make a `sqlc.yaml` similar to the following:

   ```yaml
   version: "2"

   plugins:
     - name: zz
       process:
         cmd: sqlc-gen-zombiezen

   sql:
     - engine: sqlite
       queries: "./queries"
       schema: "./migrations"
       codegen:
         - out: zz
           plugin: zz
  ```

The generator understands `sqlc.slice('param')` macros and will emit query helpers that accept Go slices, rewrite the SQL placeholders, and bind each element for you.

### Options

You can configure the plugin with the `options` block on a codegen target. For example,
to skip generating CRUD helpers and automatic `time.Time` coercion:

```yaml
sql:
  - engine: sqlite
    queries: ./queries
    schema: ./migrations
    codegen:
      - out: zz
        plugin: zz
        options:
          disable_crud: true
          disable_time_conversion: true
```

- `disable_crud` (default `false`): Skip generating CRUD helpers.
- `disable_time_conversion` (default `false`): Leave timestamp-like columns as their raw types instead of `time.Time`.

The generated Go package name is automatically derived from the final segment of the configured `out` path, so pointing `out` at `internal/drivers` yields package `drivers`.

The plugin removes the `out` directory before every run so stale files don’t linger.

To generate code into a different folder (and package), just adjust `out`:

```yaml
sql:
  - engine: sqlite
    queries: ./queries
    schema: ./migrations
    codegen:
      - out: internal/codegen
        plugin: zz
```

3. Run sqlc: `sqlc generate`
4. ???
5. Profit
