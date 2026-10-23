"""Prepare internal Go module requirements for the shared CDS release tag.

CDS determines the project version and tags every module in go.work. Its Go
plugin does not edit go.mod requirements, so this step runs before CDS tags the
release, ensuring published modules refer to the same project version.
"""

import argparse
import json
import subprocess
from pathlib import Path


ROOT = Path(__file__).resolve().parent


def go_json(*args: str, cwd: Path = ROOT) -> dict:
    return json.loads(
        subprocess.check_output(["go", *args], cwd=cwd, text=True, encoding="utf-8")
    )


def prepare(version: str, dry_run: bool = False) -> int:
    if not version or not version[0].isdigit():
        raise ValueError("Expected a CDS project version such as 1.171.0")
    tag = f"v{version}"
    module_dirs = [ROOT / entry["DiskPath"] for entry in go_json("work", "edit", "-json")["Use"]]
    module_paths = {
        go_json("mod", "edit", "-json", cwd=directory)["Module"]["Path"]: directory
        for directory in module_dirs
    }
    if len(module_paths) != len(module_dirs):
        raise ValueError("go.work contains duplicate Go module paths")
    discovered = set((ROOT / "utils").rglob("go.mod"))
    if {directory / "go.mod" for directory in module_dirs} != discovered:
        raise ValueError("go.work must include every go.mod before preparing a release")

    changed = 0
    for directory in module_dirs:
        manifest = go_json("mod", "edit", "-json", cwd=directory)
        for requirement in manifest.get("Require", []):
            path = requirement["Path"]
            if path not in module_paths or requirement["Version"] == tag:
                continue
            print(f"{directory.relative_to(ROOT)}: {path} -> {tag}")
            if not dry_run:
                subprocess.run(["go", "mod", "edit", f"-require={path}@{tag}"], cwd=directory, check=True)
            changed += 1
    print(f"Prepared {changed} internal requirements in {len(module_dirs)} CDS modules")
    return changed


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("version", help="Project version from cd-determine-version")
    parser.add_argument("--dry-run", action="store_true")
    arguments = parser.parse_args()
    prepare(arguments.version, dry_run=arguments.dry_run)
