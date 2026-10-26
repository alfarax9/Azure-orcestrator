// Package domain holds the facts Astra stores about workspaces, projects and workers, and
// the functions that turn those facts into what people see: a status from the shared
// vocabulary and the stages of the loop strip. It depends on the standard library only.
//
// The daemon stores facts, never status (PRD section 7). Status and stages are derived
// every time they are read, here and, with the same rules, in the web client; the cases in
// contract/status-cases.json keep the two implementations identical.
package domain
