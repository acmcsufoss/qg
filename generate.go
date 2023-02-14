//go:build ignore

package ignore

//go:generate jsonnet ./types/qg.jsonnet -o ./types/qg.json
//go:generate prettier -w ./types/qg.json

//go:generate jtd-codegen ./types/qg.json --go-package qg --go-out ./backend/qg/
//go:generate gofmt -w ./backend/qg/qg.go
//go:generate cp ./types/qg.json ./backend/qg/schema.json
//go:generate go generate ./backend/qg/stores/sqlite
//go:generate ./scripts/jtd-go-validators ./backend/qg/qg_validators.go

//go:generate jtd-codegen ./types/qg.json --typescript-out ./frontend/src/lib/qg-jtd/
//go:generate prettier -w ./frontend/src/lib/qg-jtd/index.ts
//go:generate ./scripts/json2ts -i jtd -t jtd.Schema ./types/qg.json ./frontend/src/lib/qg-jtd/schema.ts
