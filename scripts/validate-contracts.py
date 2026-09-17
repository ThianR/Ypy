import json
from pathlib import Path
from jsonschema import Draft202012Validator
import yaml

root = Path(__file__).resolve().parents[1]
checks = [
    (root / "contracts/events/pos-operation.schema.json", root / "contracts/fixtures/pos-operation.json"),
    (root / "contracts/events/envelope.schema.json", root / "contracts/fixtures/event-envelope.json"),
    (root / "contracts/events/fiscal-document.schema.json", root / "contracts/fixtures/fiscal-document.json"),
]
for schema_path, fixture_path in checks:
    schema = json.loads(schema_path.read_text(encoding="utf-8"))
    fixture = json.loads(fixture_path.read_text(encoding="utf-8"))
    errors = sorted(Draft202012Validator(schema).iter_errors(fixture), key=lambda e: list(e.path))
    if errors:
        for error in errors: print(f"{fixture_path.name} {list(error.path)}: {error.message}")
        raise SystemExit(1)
    print(f"{fixture_path.name} validates against {schema_path.name}.")

openapi = yaml.safe_load((root / "contracts/http/openapi.yaml").read_text(encoding="utf-8"))
if not isinstance(openapi, dict) or openapi.get("openapi") != "3.0.3" or not isinstance(openapi.get("paths"), dict) or not isinstance(openapi.get("components", {}).get("schemas"), dict):
    raise SystemExit("openapi.yaml no contiene una estructura OpenAPI 3.0.3 válida")
for required_path in ("/api/v1/auth/login", "/api/v1/pos/operations", "/api/v1/catalog/scale-rules", "/api/v1/catalog/barcodes", "/api/v1/catalog/items", "/api/v1/catalog/prices", "/api/v1/maestros/quotes", "/api/v1/cash/open", "/api/v1/cash/close"):
    if required_path not in openapi["paths"]:
        raise SystemExit(f"Falta ruta OpenAPI: {required_path}")
print("openapi.yaml structure validated.")
