# VietGara Test Data

> **Purpose:** This repo is the **shared test dataset** of VietGara: one YAML description of the users, garages, staff, customers, suppliers and parts that the backend blackbox suite and the web end-to-end suite create, use and remove. It contains data and a validator, no application code.

| | |
| --- | --- |
| **Type** | Test data (YAML + JSON Schema), private repository |
| **Used by** | `vietgara-backend` (`testing/blackbox`), `vietgara-e2e`, and `vietgara-docs` (the generated `00_Test_Data.md`) |
| **Does** | Single source of truth for named test personas and records, so the documents, the API tests and the browser tests never drift apart |
| **Built with** | YAML, JSON Schema (draft 2020-12), a Python validator; Go and TypeScript loaders are added in Release 4.0 |
| **Depends on** | Nothing at run time. A seeder (Release 4.0) creates the data through the public VietGara API |
| **Related** | `vietgara-docs` (Release Plan Phase 4, Test Strategy, E2E Scenarios), `vietgara-backend`, `vietgara-e2e`, `vietgara-infra` |

Status: **initial dataset.** The structure, naming and baseline records are in place and validated. The Go and TypeScript loaders, the seeder and the test-only purge are built in Release 4.0 (Phase 4, Quality Foundation) of the [Release Plan](https://github.com/viettechno/vietgara-docs/blob/master/docs/01-Product/Release_Plan/Release_Plan_Phase4.md).

## Layout

```
schema/dataset.schema.json   JSON Schema of the merged dataset
data/                        One YAML file per area; the top-level keys are merged
  environments.yaml          E-mail domain and how the password is supplied per target
  plans.yaml                 Subscription plans the tests create
  users.yaml                 Platform users and owners, staff, freelancer, unverified user
  garages.yaml               Garages, custom roles and groups, staff memberships
  customers.yaml             Customers with their vehicles
  suppliers.yaml             Suppliers (not customers, not insurers)
  catalogue.yaml             Units, part categories and parts
scripts/validate.py          Merges data/, checks the schema and the references
```

## Names

Short, predictable names. The same key is the user name part of the e-mail, so a tester can type it from memory.

| Thing | Pattern | Examples |
| --- | --- | --- |
| Owner | `owner<n>` | `owner1`, `owner2` |
| Garage | `gara<owner>-<n>` | `gara1-1`, `gara1-2` (first and second garage of owner1), `gara2-1` |
| Staff | `<role><owner>-<garage>` | `manager1-1`, `advisor1-1`, `technician1-1`, `inventory1-1`, `accountant1-1` |
| Customer | `customer<owner>-<garage>-<n>` | `customer1-1-1` |
| Supplier | `supplier<owner>-<garage>-<n>` | `supplier1-1-1` |
| Platform users | `admin`, `support<n>` | `admin` (Platform Admin), `support1` (admin staff) |
| Single-use users | created by a scenario with a unique suffix | `signup.iam001.<unique>` |

E-mail addresses always use the test domain (`<key>@vietgara.test`). **Never put a real mail domain in this repository:** staging sends real e-mail, and the test domain is routed to the test inbox (Mailpit).

## Two kinds of data

- **Baseline data** (everything in `data/`) has fixed names and is read-mostly. A test must not leave it changed.
- **Scenario-owned data** is created by the test itself with a unique suffix (sign-up, password reset, destructive scenarios) and deleted in teardown. It is not described here.

## Passwords and secrets

- On the local and stubbed stacks every baseline user has the one password in `data/environments.yaml` (`stubbed`).
- On staging the dataset holds only `passwordRef: TEST_PASSWORD`; the value is a CI secret, the same for every persona. The platform admin has its own secret (`ADMIN_PASSWORD`). When MFA ships, test admins get a fixed TOTP secret from a secret too.
- **No real credential, token or personal data is ever committed.** The repository is private because staging is on the internet and a predictable name with a known password would be an open door.

## Clean-up

The API cannot remove used records (the delete-or-deactivate rule keeps them), so after a run a privileged, test-only purge removes every user under the test domain with their garages, files, outbox rows and Mailpit messages. It is disabled in production by configuration and refuses anything outside the test domain. The stubbed stack is simply torn down with its volumes. Both are built in Release 4.0.

## Use it

```bash
pip install pyyaml jsonschema
make validate        # merge data/, check the schema, check every reference
```

The consuming suites pin a **tag** of this repository (`v0.1.0`, ...). Change the dataset by pull request; a change that renames or removes a persona is a breaking change and bumps the minor version while the version is below 1.0.

## Sections that wait for a feature

Some records need a feature that arrives in a later phase. They carry `requires`, and the seeder skips a section whose requirement the target does not offer yet.

| `requires` | Feature | Release Plan |
| --- | --- | --- |
| `phase-3` | `user` terminology (the dataset itself already uses it) | Phase 3 |
| `phase-5` | Suppliers table, vouchers, tax profiles | Release 5.0 / 5.1 |
| `phase-6` | Units, part categories, MFA | Release 6.0 / 6.1 |
| `phase-7` | Platform staff and roles | Release 7.1 |
