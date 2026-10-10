import { readdir, readFile } from "node:fs/promises";
import { join } from "node:path";
import { parseAllDocuments } from "yaml";

export interface Dataset {
  version: number;
  emailDomain: string;
  environments: Record<string, Environment>;
  plans: Plan[];
  users: User[];
  garages: Garage[];
  roles?: Role[];
  groups?: Group[];
  staff: Staff[];
  customers: Customer[];
  suppliers: Supplier[];
  catalogue?: Catalogue;
}

export interface Environment {
  password?: string;
  adminPassword?: string;
  passwordRef?: string;
  adminPasswordRef?: string;
}

export interface Plan {
  code: string;
  name: string;
  price: number;
  currency: string;
  billingCycle: "monthly" | "yearly";
  limits: Limits;
  features?: string[];
}

export interface Limits {
  maxGarages: number;
  maxStaff: number;
  maxTransactionsPerDay: number;
}

export interface User {
  key: string;
  name: string;
  phone?: string;
  plan?: string;
  platformAdmin?: boolean;
  platformRole?: string;
  supplierFor?: string;
  verified?: boolean;
  requires?: string;
  note?: string;
}

export interface Garage {
  key: string;
  owner: string;
  name: string;
  address?: string;
  phone?: string;
  taxCode?: string;
}

export interface Role {
  owner: string;
  name: string;
  permissions: string[];
}

export interface Group {
  garage: string;
  name: string;
  role: string;
}

export interface Staff {
  user: string;
  garage: string;
  group: string;
}

export interface Customer {
  key: string;
  garage: string;
  name: string;
  phone?: string;
  email?: string;
  consent?: boolean;
  vehicles?: Vehicle[];
}

export interface Vehicle {
  plate: string;
  make: string;
  model: string;
  year: number;
}

export interface Supplier {
  key: string;
  garage: string;
  name: string;
  taxCode?: string;
  contact?: SupplierContact;
}

export interface SupplierContact {
  person?: string;
  phone?: string;
  email?: string;
}

export interface Catalogue {
  requires?: string;
  garage: string;
  units: Unit[];
  categories?: Category[];
  items: Item[];
}

export interface Unit {
  key: string;
  name: string;
  abbreviation: string;
}

export interface Category {
  key: string;
  name: string;
  children?: Category[];
}

export interface Item {
  sku: string;
  name: string;
  category?: string;
  brand?: string;
  baseUnit: string;
  packaging?: Packaging[];
  reorderThreshold?: number;
  defaultSupplier?: string;
}

export interface Packaging {
  unit: string;
  factor: number;
}

const datasetSections = new Set<keyof Dataset>([
  "version",
  "emailDomain",
  "environments",
  "plans",
  "users",
  "garages",
  "roles",
  "groups",
  "staff",
  "customers",
  "suppliers",
  "catalogue",
]);

/**
 * Loads and merges the YAML files in dataDir into the shared typed dataset.
 * Duplicate top-level sections and unknown top-level keys are rejected.
 * Use the repository's make validate target for schema and reference checks.
 */
export async function load(dataDir: string): Promise<Dataset> {
  let entries;
  try {
    entries = await readdir(dataDir, { withFileTypes: true });
  } catch (error) {
    throw new Error(`read dataset directory "${dataDir}": ${errorMessage(error)}`, { cause: error });
  }

  const yamlFiles = entries
    .filter((entry) => entry.isFile() && entry.name.endsWith(".yaml"))
    .sort((left, right) => left.name.localeCompare(right.name));

  if (yamlFiles.length === 0) {
    throw new Error(`dataset directory "${dataDir}" contains no .yaml files`);
  }

  const merged: Record<string, unknown> = Object.create(null) as Record<string, unknown>;

  for (const file of yamlFiles) {
    const path = join(dataDir, file.name);
    let content: string;
    try {
      content = await readFile(path, "utf8");
    } catch (error) {
      throw new Error(`read dataset file "${path}": ${errorMessage(error)}`, { cause: error });
    }

    const documents = parseAllDocuments(content, { uniqueKeys: true });
    const yamlError = documents.flatMap((document) => document.errors)[0];
    if (yamlError) {
      throw new Error(`parse dataset file "${path}": ${yamlError.message}`, { cause: yamlError });
    }
    const document = documents[0];
    if (documents.length !== 1 || document?.contents === null || document === undefined) {
      throw new Error(`parse dataset file "${path}": expected one YAML mapping`);
    }

    const value: unknown = document.toJS();
    if (!isRecord(value)) {
      throw new Error(`parse dataset file "${path}": expected one YAML mapping`);
    }

    for (const [section, contents] of Object.entries(value)) {
      if (!datasetSections.has(section as keyof Dataset)) {
        throw new Error(`dataset file "${path}": unknown top-level field "${section}"`);
      }
      if (Object.hasOwn(merged, section)) {
        throw new Error(`dataset section "${section}" is defined in more than one file`);
      }
      merged[section] = contents;
    }
  }

  return merged as unknown as Dataset;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}
