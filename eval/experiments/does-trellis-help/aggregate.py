#!/usr/bin/env python3
"""Roll per-run reviewer scores into the baseline-vs-+Trellis Δ (research-0011).

  python3 eval/experiments/does-trellis-help/aggregate.py runs/

Reads runs/<framework>/<task>/<arm>-<idx>.<rubric>.score.md — the reviewer's output, whose
last line is `SUMMARY | followed=<n> violated=<n> n-a=<n>` (falls back to counting verdict
lines). Prints, per rubric, the mean followed/violated for each arm and the Δ (trellis − baseline).
"""
import re
import sys
import pathlib
from collections import defaultdict

root = pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else str(pathlib.Path(__file__).parent / "runs"))
# tallies[rubric][arm] = [followed_list, violated_list]
tallies = defaultdict(lambda: defaultdict(lambda: ([], [])))

for f in root.rglob("*.score.md"):
    m = re.match(r"(baseline|trellis)-\d+\.(.+)\.score$", f.name[:-3])
    if not m:
        continue
    arm, rubric = m.group(1), m.group(2)
    text = f.read_text()
    s = re.search(r"SUMMARY \| followed=(\d+) violated=(\d+)", text)
    if s:
        followed, violated = int(s.group(1)), int(s.group(2))
    else:  # fallback: count verdict lines
        followed = len(re.findall(r"\| followed \|", text))
        violated = len(re.findall(r"\| violated \|", text))
    tallies[rubric][arm][0].append(followed)
    tallies[rubric][arm][1].append(violated)


def mean(xs):
    return sum(xs) / len(xs) if xs else 0.0


if not tallies:
    print(f"no *.score.md files under {root}/", file=sys.stderr)
    sys.exit(1)

print(f"{'rubric':<24} {'arm':<9} {'n':>3} {'followed':>9} {'violated':>9}")
print("-" * 60)
for rubric in sorted(tallies):
    base_f = base_v = None
    for arm in ("baseline", "trellis"):
        fol, vio = tallies[rubric][arm]
        if not fol:
            continue
        print(f"{rubric:<24} {arm:<9} {len(fol):>3} {mean(fol):>9.1f} {mean(vio):>9.1f}")
        if arm == "baseline":
            base_f, base_v = mean(fol), mean(vio)
        elif base_f is not None:
            df, dv = mean(fol) - base_f, mean(vio) - base_v
            print(f"{rubric:<24} {'Δ':<9} {'':>3} {df:>+9.1f} {dv:>+9.1f}"
                  f"   (+followed / −violated ⇒ Trellis helps)")
    print()
