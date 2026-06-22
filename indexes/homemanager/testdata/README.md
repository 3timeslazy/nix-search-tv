# Home Manager test fixtures

`TestFetcher` cross-checks the HTML parser (`ParseMdBook`, fed `print.html`)
against the canonical option set from `nix build`'d `options.json`. The two
fixtures below must therefore be generated from the **same** Home Manager
revision.

## Regenerating

```sh
# 1. options.json (canonical option set, from the docs-json output)
nix build github:nix-community/home-manager/master#docs-json \
  --no-link --print-out-paths
#   -> <out>/share/doc/home-manager/options.json

# 2. print.html (the single-page mdBook rendering parsed at runtime)
nix build github:nix-community/home-manager/master#docs-html \
  --no-link --print-out-paths
#   -> <out>/share/doc/home-manager/print.html
```

Copy `print.html` verbatim. `options.json` needs one normalisation step
(below) before it is copied in.

## options.json normalisation: `declarations`

Most options list their `declarations` as `{ "name", "url" }` objects, which
decode straight into `homemanager.Declarations`. A handful of options coming
from the shared `modules/generic/meta-maintainers.nix` (the `meta.maintainers`
/ `meta.teams` family, including the nested copies under
`programs.<browser>.profiles.<name>.*`) instead emit a **bare string** store
path. The docs HTML renders that same path as a `file://` link, so to keep the
two fixtures comparable we rewrite those string declarations into the object
form with a `file://` URL:

```sh
jq '
  def fix:
    if type == "string"
    then { name: ("<home-manager/" + (sub("^/nix/store/[^/]+/"; "")) + ">"),
           url: ("file://" + .) }
    else . end;
  map_values(if .declarations then .declarations |= map(fix) else . end)
' options.json > options.normalised.json && mv options.normalised.json options.json
```

After this, no `declarations` entry should be a plain string:

```sh
jq '[ .[].declarations[]? | select(type == "string") ] | length' options.json  # -> 0
```
