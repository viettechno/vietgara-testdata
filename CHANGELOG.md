# Changelog

## Unreleased

## 0.3.0 — Release 6.0

- Selling prices (`salePrice`, per base unit; `packaging[].salePrice` overrides base price × factor) in the schema, the Go and TypeScript loaders and the catalogue data.
- New category `beverages` and part `BEER-330` (base unit can, case = 24 priced 200,000), the Definition-of-Done part of Release 6.0.

- Add a TypeScript loader package for merging the shared YAML dataset.
- Add a typed Go loader for merging the shared YAML dataset.
- Add the supplier user `supplieruser1` (supplier portal actor, `requires: phase-17`); platform staff now `requires: phase-8` after Phases 7 and 8 swapped in the Release Plan.

## 0.1.0 — initial dataset

- Structure, naming and the baseline of Release 4.0: plans, platform users, two owners with three garages, staff, customers with vehicles, suppliers, units, categories and parts.
- JSON Schema and a validator (`make validate`).
