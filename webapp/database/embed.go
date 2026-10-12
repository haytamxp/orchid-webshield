package database

import _ "embed"

// Schema contient le texte de schema.sql, intégré au programme à la compilation.
//
//go:embed schema.sql
var Schema string