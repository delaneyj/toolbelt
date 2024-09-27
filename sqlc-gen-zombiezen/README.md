# SQLC Zombiezen driver plugin

## How to use

1. Install the plugin:
   ```shell
   go install github.com/delaneyj/toolbelt/sqlc-gen-zombiezen@latest
   ```
2. Make a sqlc.yml similar to the following:

   ```yaml
   version: "2"

   plugins:
   - name: zz
       process:
       cmd: sqlc-gen-zombiezen

   sql:
   - engine: "sqlite"
       queries: "./queries"
       schema: "./migrations"
       codegen:
       - out: zz
           plugin: zz
   ```

3. Run sqlc: `sqlc generate`
4. ???
5. Profit
