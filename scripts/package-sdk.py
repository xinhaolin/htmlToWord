"""Build a portable source archive; never include local document/client artifacts."""
from pathlib import Path
import argparse, hashlib, json, zipfile

ROOT = Path(__file__).resolve().parents[1]
p = argparse.ArgumentParser()
p.add_argument('--output', default=str(ROOT / 'dist/htmlword-sdk.zip'))
args = p.parse_args()
destination = Path(args.output).resolve()
destination.parent.mkdir(parents=True, exist_ok=True)
allowed_dirs = {'.github', 'cmd', 'docs', 'examples', 'internal', 'scripts', 'testdata', 'tools'}
excluded_parts = {'.git', '__pycache__', 'artifacts', 'bin', 'dist', 'node_modules'}
files = []
for path in sorted(ROOT.rglob('*')):
    if not path.is_file() or path.is_symlink():
        continue
    rel = path.relative_to(ROOT)
    if any(part in excluded_parts for part in rel.parts):
        continue
    if len(rel.parts) > 1 and rel.parts[0] not in allowed_dirs:
        continue
    if path.name == '.DS_Store' or path.suffix in {'.docx', '.pdf', '.zip', '.pyc', '.log'}:
        continue
    if len(rel.parts) == 1 and path.suffix == '.png':
        continue
    if path.name in {'verification.json', 'start.command'}:
        continue
    if path.suffix == '.md' and path.name in {'legacy-workbench.zh-CN.md', 'validation-history.zh-CN.md'}:
        continue
    files.append((path, rel))
with zipfile.ZipFile(destination, 'w', zipfile.ZIP_DEFLATED) as archive:
    for path, rel in files:
        archive.write(path, 'htmlword/' + rel.as_posix())
digest = hashlib.sha256(destination.read_bytes()).hexdigest()
print(json.dumps({'archive': str(destination), 'files': len(files), 'sha256': digest}, indent=2))
