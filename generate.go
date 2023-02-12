package main

//go:generate ./types/qg/qg-update
//go:generate jsonnet ./types/qg.jsonnet -o ./types/qg.json
//go:generate jtd-codegen ./types/qg.json --go-package qg --go-out ./backend/qg/
//go:generate jtd-codegen ./types/qg.json --typescript-out ./frontend/src/lib/qg-jtd/
//go:generate ./scripts/json2js ./types/qg.json ./frontend/src/lib/qg-jtd/schema.js
//go:generate cp ./types/qg.json ./backend/qg/schema.json
