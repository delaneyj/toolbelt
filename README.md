# toolbelt
A set of utilities used in every go project

<img src="./assets/toolbelt.png" width="512" height="512" alt="wisshes mascot"/>

## Package splits
Several helpers were moved out of the root `toolbelt` package to avoid pulling external dependencies into every import.

- `toolbelt/db`: sqlite + timestamp helpers from `database.go` (Database, NewDatabase, Julian/Stmt helpers, migrations).
- `toolbelt/protobuf`: protobuf marshal/unmarshal helpers from `protobuf.go`.
- `toolbelt/egctx`: errgroup helpers from `egctx.go`.
- `toolbelt/id`: ID generation/encoding helpers from `id.go` (no chi dependency).
- `toolbelt/web`: chi-based request param helpers (ChiParamInt64, ChiParamEncodedID).

If you were importing these from `toolbelt` directly, update your imports to the package listed above.
