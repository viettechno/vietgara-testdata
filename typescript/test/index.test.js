import assert from "node:assert/strict";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import { load } from "../dist/index.js";

async function withDataDir(t, files) {
  const dataDir = await mkdtemp(path.join(os.tmpdir(), "vietgara-testdata-"));
  t.after(() => rm(dataDir, { recursive: true, force: true }));

  for (const [name, contents] of Object.entries(files)) {
    await writeFile(path.join(dataDir, name), contents);
  }
  return dataDir;
}

test("load merges files into the typed dataset", async (t) => {
  const dataDir = await withDataDir(t, {
    "environments.yaml": `version: 1
emailDomain: vietgara.test
environments:
  stubbed:
    password: local-password
    adminPassword: local-admin-password
`,
    "users.yaml": `users:
  - key: owner1
    name: QA Owner
    verified: true
`,
    "catalogue.yaml": `catalogue:
  requires: phase-6
  garage: gara1-1
  units:
    - key: litre
      name: Litre
      abbreviation: L
  items:
    - sku: OIL-5W30
      name: Engine oil
      baseUnit: litre
      packaging:
        - unit: can
          factor: 4
`,
  });

  const dataset = await load(dataDir);
  assert.equal(dataset.version, 1);
  assert.equal(dataset.emailDomain, "vietgara.test");
  assert.equal(dataset.environments.stubbed.password, "local-password");
  assert.deepEqual(dataset.users[0], {
    key: "owner1",
    name: "QA Owner",
    verified: true,
  });
  assert.equal(dataset.catalogue.items[0].packaging[0].factor, 4);
});

test("load reads the repository dataset", async () => {
  const dataDir = fileURLToPath(new URL("../../data/", import.meta.url));
  const dataset = await load(dataDir);

  assert.equal(dataset.version, 1);
  assert.equal(dataset.users.length, 13);
  assert.equal(dataset.garages.length, 3);
});

test("load rejects invalid dataset files", async (t) => {
  const cases = [
    {
      name: "no yaml files",
      files: { "notes.txt": "not a dataset" },
      message: "contains no .yaml files",
    },
    {
      name: "empty yaml file",
      files: { "empty.yaml": "" },
      message: "expected one YAML mapping",
    },
    {
      name: "non-mapping document",
      files: { "data.yaml": "- item" },
      message: "expected one YAML mapping",
    },
    {
      name: "malformed yaml",
      files: { "data.yaml": "users: [" },
      message: "parse dataset file",
    },
    {
      name: "duplicate section in one file",
      files: { "data.yaml": "users: []\nusers: []\n" },
      message: "Map keys must be unique",
    },
    {
      name: "duplicate section across files",
      files: {
        "a.yaml": "users: []\n",
        "b.yaml": "users: []\n",
      },
      message: 'dataset section "users" is defined in more than one file',
    },
    {
      name: "unknown top-level field",
      files: { "data.yaml": "usres: []\n" },
      message: 'unknown top-level field "usres"',
    },
    {
      name: "multiple documents",
      files: { "data.yaml": "users: []\n---\nusers: []\n" },
      message: "expected one YAML mapping",
    },
  ];

  for (const { name, files, message } of cases) {
    await t.test(name, async (subtest) => {
      const dataDir = await withDataDir(subtest, files);
      await assert.rejects(
        load(dataDir),
        (error) => error instanceof Error && error.message.includes(message),
      );
    });
  }
});

test("load reports directory read errors", async (t) => {
  const missingDir = path.join(os.tmpdir(), `vietgara-missing-${process.pid}`);
  await assert.rejects(load(missingDir), (error) =>
    error instanceof Error && error.message.includes(`read dataset directory "${missingDir}"`),
  );
});
