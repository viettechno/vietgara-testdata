#!/usr/bin/env python3
"""Merge data/*.yaml, validate it against schema/dataset.schema.json and check references.

Exit status 0 when the dataset is valid, 1 otherwise. Needs: pip install pyyaml jsonschema
"""
import json
import pathlib
import re
import sys

try:
    import yaml
    from jsonschema import Draft202012Validator
except ImportError:  # pragma: no cover
    sys.exit("missing dependency: pip install pyyaml jsonschema")

ROOT = pathlib.Path(__file__).resolve().parent.parent


def load():
    merged = {}
    for f in sorted((ROOT / "data").glob("*.yaml")):
        doc = yaml.safe_load(f.read_text()) or {}
        for key, value in doc.items():
            if key in merged:
                sys.exit(f"{f.name}: top-level key '{key}' is defined in two files")
            merged[key] = value
    return merged


def categories(nodes):
    for n in nodes or []:
        yield n["key"]
        yield from categories(n.get("children"))


def main():
    data = load()
    schema = json.loads((ROOT / "schema" / "dataset.schema.json").read_text())
    errors = [f"schema: {'/'.join(map(str, e.absolute_path)) or '(root)'}: {e.message}"
              for e in sorted(Draft202012Validator(schema).iter_errors(data), key=lambda e: list(map(str, e.absolute_path)))]

    users = [u["key"] for u in data.get("users", [])]
    garages = [g["key"] for g in data.get("garages", [])]
    owners = {g["owner"] for g in data.get("garages", [])}
    plans = {p["code"] for p in data.get("plans", [])}
    builtin_plans = {"trial", "basic", "pro"}
    customers = [c["key"] for c in data.get("customers", [])]
    suppliers = [s["key"] for s in data.get("suppliers", [])]

    def dup(kind, keys):
        for k in {k for k in keys if keys.count(k) > 1}:
            errors.append(f"{kind} '{k}' is defined more than once")

    dup("user", users); dup("garage", garages); dup("customer", customers); dup("supplier", suppliers)

    for g in data.get("garages", []):
        if g["owner"] not in users:
            errors.append(f"garage {g['key']}: owner '{g['owner']}' is not a user")
        elif not re.fullmatch(rf"gara{re.escape(g['owner'].removeprefix('owner'))}-[0-9]+", g["key"]):
            errors.append(f"garage {g['key']}: the key must be gara<owner number>-<n> for owner {g['owner']}")
    for u in data.get("users", []):
        plan = u.get("plan")
        if plan and plan not in plans | builtin_plans:
            errors.append(f"user {u['key']}: unknown plan '{plan}'")
        if plan and u["key"] not in owners:
            errors.append(f"user {u['key']}: has a plan but owns no garage")
    group_names = {"Manager", "Accountant", "Service Advisor", "Technician", "Inventory Staff"}
    for gr in data.get("groups", []):
        if gr["garage"] not in garages:
            errors.append(f"group {gr['name']}: unknown garage '{gr['garage']}'")
        group_names.add(gr["name"])
    role_names = {r["name"] for r in data.get("roles", [])}
    for gr in data.get("groups", []):
        if gr["role"] not in role_names:
            errors.append(f"group {gr['name']}: unknown role '{gr['role']}'")
    for r in data.get("roles", []):
        if r["owner"] not in owners:
            errors.append(f"role {r['name']}: '{r['owner']}' owns no garage")
    seen = set()
    for s in data.get("staff", []):
        if s["user"] not in users:
            errors.append(f"staff: unknown user '{s['user']}'")
        if s["garage"] not in garages:
            errors.append(f"staff {s['user']}: unknown garage '{s['garage']}'")
        if s["group"] not in group_names:
            errors.append(f"staff {s['user']}: unknown group '{s['group']}'")
        if (s["user"], s["garage"]) in seen:
            errors.append(f"staff {s['user']} appears twice in {s['garage']}")
        seen.add((s["user"], s["garage"]))
    for c in data.get("customers", []):
        if c["garage"] not in garages:
            errors.append(f"customer {c['key']}: unknown garage '{c['garage']}'")
        if not c["key"].startswith("customer" + c["garage"].removeprefix("gara") + "-"):
            errors.append(f"customer {c['key']}: the key must start with customer<owner>-<garage> of {c['garage']}")
    plates = {}
    for c in data.get("customers", []):
        for v in c.get("vehicles", []):
            if (c["garage"], v["plate"]) in plates:
                errors.append(f"vehicle {v['plate']} is defined twice in {c['garage']}")
            plates[(c["garage"], v["plate"])] = c["key"]
    for s in data.get("suppliers", []):
        if s["garage"] not in garages:
            errors.append(f"supplier {s['key']}: unknown garage '{s['garage']}'")
        if not s["key"].startswith("supplier" + s["garage"].removeprefix("gara") + "-"):
            errors.append(f"supplier {s['key']}: the key must start with supplier<owner>-<garage> of {s['garage']}")

    cat = data.get("catalogue")
    if cat:
        if cat["garage"] not in garages:
            errors.append(f"catalogue: unknown garage '{cat['garage']}'")
        units = {u["key"] for u in cat["units"]}
        cats = set(categories(cat.get("categories")))
        skus = [i["sku"] for i in cat["items"]]
        dup("sku", skus)
        for i in cat["items"]:
            if i["baseUnit"] not in units:
                errors.append(f"item {i['sku']}: unknown base unit '{i['baseUnit']}'")
            for p in i.get("packaging", []):
                if p["unit"] not in units:
                    errors.append(f"item {i['sku']}: unknown packaging unit '{p['unit']}'")
            if i.get("category") and i["category"] not in cats:
                errors.append(f"item {i['sku']}: unknown category '{i['category']}'")
            sup = i.get("defaultSupplier")
            if sup and sup not in suppliers:
                errors.append(f"item {i['sku']}: unknown supplier '{sup}'")
            elif sup and next(s for s in data["suppliers"] if s["key"] == sup)["garage"] != cat["garage"]:
                errors.append(f"item {i['sku']}: supplier '{sup}' belongs to another garage")

    if errors:
        print("\n".join(errors))
        print(f"\n{len(errors)} problem(s)")
        return 1
    print(f"dataset OK: {len(users)} users, {len(garages)} garages, {len(data.get('staff', []))} staff memberships, "
          f"{len(customers)} customers, {len(suppliers)} suppliers, {len(cat['items']) if cat else 0} parts")
    return 0


if __name__ == "__main__":
    sys.exit(main())
