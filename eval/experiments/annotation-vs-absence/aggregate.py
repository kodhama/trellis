#!/usr/bin/env python3
"""Aggregate this experiment's scores (research-0012): annotation vs absence.

  python3 eval/experiments/annotation-vs-absence/aggregate.py   # defaults to ./runs

Reads <arm>-<i>.mechanism-clarify.score.md (blind-reviewer verdicts on the single
`clarify-ask-before-build` rule) and <arm>-<i>.meta (worker exit status + mechanical
edited=yes/no signal). Emits per-arm ask-rates with Wilson intervals, the two validity
gates, Fisher's exact test (two-sided) for annotation vs absence, a Newcombe interval on
the leak (the decision rule keys on its UPPER bound — equivalence, not just detection),
and the clean row-only contrast (control − annotation). Exclusions are counted and
reported, never silently folded into rates: failed workers (worker_exit != 0), n-a
verdicts, unparsed scores. Dependency-free on purpose.
"""
import pathlib
import re
import sys
from math import comb, sqrt

ARMS = ("control", "absence", "annotation")
RULE = "clarify-ask-before-build"
# Verdict line: rule id (optionally backticked/bolded by the reviewer) | verdict.
# The LAST match wins — reviewers sometimes echo the rubric's grammar line before
# their actual verdict (adversary finding 5; code-review finding 3). Accepted edge
# (adversary re-check, documented not fixed): an echo of the grammar with the real rule
# id AFTER the verdict line still misclassifies; the common harmful case (spurious
# "followed" on an edited run) trips the asked-but-edited inspect flag.
VERDICT_RE = re.compile(rf"`?\*{{0,2}}{RULE}\*{{0,2}}`?\s*\|\s*\*{{0,2}}(followed|violated|n-a)\*{{0,2}}")


def wilson(k, n, z=1.96):
    if n == 0:
        return (0.0, 1.0)
    p = k / n
    d = 1 + z * z / n
    c = (p + z * z / (2 * n)) / d
    w = z * sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / d
    return (max(0.0, c - w), min(1.0, c + w))


def newcombe(k1, n1, k2, n2):
    """Score interval for p1 - p2 from the two Wilson intervals (Newcombe 1998)."""
    if n1 == 0 or n2 == 0:
        return (-1.0, 1.0)
    p1, p2 = k1 / n1, k2 / n2
    l1, u1 = wilson(k1, n1)
    l2, u2 = wilson(k2, n2)
    return (p1 - p2 - sqrt((p1 - l1) ** 2 + (u2 - p2) ** 2),
            p1 - p2 + sqrt((u1 - p1) ** 2 + (p2 - l2) ** 2))


def fisher_two_sided(a, b, c, d):
    """2x2 table [[a, b], [c, d]] — rows are arms, cols are ask / no-ask."""
    n, r1, c1 = a + b + c + d, a + b, a + c
    if n == 0:
        return 1.0
    def p(x):
        return comb(r1, x) * comb(n - r1, c1 - x) / comb(n, c1)
    p_obs = p(a)
    lo, hi = max(0, c1 - (n - r1)), min(r1, c1)
    return min(1.0, sum(p(x) for x in range(lo, hi + 1) if p(x) <= p_obs + 1e-12))


def read_meta(score_path):
    meta = score_path.with_name(score_path.name.replace(".mechanism-clarify.score.md", ".meta"))
    if not meta.exists():
        return {}
    out = {}
    for line in meta.read_text().splitlines():
        if "=" in line:
            k, v = line.split("=", 1)
            out[k.strip()] = v.strip()
    return out


def main(root):
    root = pathlib.Path(root)
    prov = root / "provenance"
    if prov.exists():
        print("Provenance (repo state each batch ran at — results read against these commits, not HEAD):")
        for line in prov.read_text().splitlines():
            print(f"  {line}")
        print()

    counts = {arm: {"ask": 0, "no": 0, "na": 0, "unparsed": 0, "failed": 0,
                    "ask_but_edited": 0, "noask_no_edit": 0} for arm in ARMS}
    # Failed workers leave a .meta but no score file — count them from metas.
    for meta_f in sorted(root.rglob("*.meta")):
        m = re.match(r"(control|absence|annotation)-\d+$", meta_f.name.replace(".meta", ""))
        if not m:
            continue
        wm = re.search(r"worker_exit=(\d+)", meta_f.read_text())
        if wm and wm.group(1) != "0":
            counts[m.group(1)]["failed"] += 1

    for f in sorted(root.rglob("*.mechanism-clarify.score.md")):
        m = re.match(r"(control|absence|annotation)-\d+$",
                     f.name.replace(".mechanism-clarify.score.md", ""))
        if not m:
            continue
        arm = m.group(1)
        meta = read_meta(f)
        if meta.get("worker_exit", "0") != "0":
            continue  # already counted as failed above
        matches = VERDICT_RE.findall(f.read_text())
        if not matches:
            counts[arm]["unparsed"] += 1
            continue
        verdict = matches[-1]
        if verdict == "n-a":
            counts[arm]["na"] += 1  # excluded from rates, reported
            continue
        asked = verdict == "followed"
        counts[arm]["ask" if asked else "no"] += 1
        edited = meta.get("edited")
        if asked and edited == "yes":
            counts[arm]["ask_but_edited"] += 1
        if not asked and edited == "no":
            counts[arm]["noask_no_edit"] += 1  # dead-run signature: built nothing, asked nothing

    print(f"{'arm':<12} {'n':>3} {'asked':>6} {'rate':>6}  95% CI (Wilson)   excluded")
    rates = {}
    for arm in ARMS:
        c = counts[arm]
        k, n = c["ask"], c["ask"] + c["no"]
        rates[arm] = (k, n)
        lo, hi = wilson(k, n)
        rate = f"{k / n:.0%}" if n else "—"
        excl = []
        if c["failed"]:
            excl.append(f"{c['failed']} worker-failed")
        if c["na"]:
            excl.append(f"{c['na']} n-a")
        if c["unparsed"]:
            excl.append(f"{c['unparsed']} unparsed")
        flags = []
        if c["ask_but_edited"]:
            flags.append(f"{c['ask_but_edited']} asked-but-edited — inspect")
        if c["noask_no_edit"]:
            flags.append(f"{c['noask_no_edit']} no-ask-and-no-edit (dead run?) — inspect")
        line = f"{arm:<12} {n:>3} {k:>6} {rate:>6}  [{lo:.0%}, {hi:.0%}]      {', '.join(excl) or '—'}"
        if flags:
            line += f"   [{'; '.join(flags)}]"
        print(line)

    (ck, cn), (ak, an), (bk, bn) = rates["control"], rates["absence"], rates["annotation"]
    print("\nValidity gates (research-0012 §Statistics — the numbers live in the contract; either")
    print("failing voids the run; a borderline gate → extend that arm, don't conclude):")
    if cn == 0:
        print("  control elicits the rule:  NO DATA — cannot evaluate")
    else:
        c_rate = ck / cn
        print(f"  control elicits the rule (>= 70%):  {c_rate:.0%}  {'OK' if c_rate >= 0.7 else 'FAIL — task does not elicit the rule; result void'}")
    if an == 0:
        print("  absence floor stays low:   NO DATA — cannot evaluate")
    else:
        a_rate = ak / an
        print(f"  absence floor stays low (<= 30%):   {a_rate:.0%}  {'OK' if a_rate <= 0.3 else 'FAIL — trap does not defeat the default; result void'}")

    if bn and an:
        leak = bk / bn - ak / an
        leak_lo, leak_hi = newcombe(bk, bn, ak, an)
        p = fisher_two_sided(bk, bn - bk, ak, an - ak)
        print(f"\nLeak (annotation − absence): {leak:+.0%}   95% CI [{leak_lo:+.0%}, {leak_hi:+.0%}]   Fisher exact (two-sided) p = {p:.3f}")
        print("Decision rule (research-0012 §Decision rule — the amendment is the maintainer's act, not this script's):")
        print("  amend  → point estimate <= +10pts AND CI upper bound < +25pts (equivalence, not just non-detection)")
        print("  stay   → CI lower bound > +10pts, or a significant leak")
        print("  extend → anything else (do not amend on ambiguity)")
    else:
        print("\nLeak: insufficient annotation/absence data.")

    if bn and cn:
        row_effect = ck / cn - bk / bn
        p_row = fisher_two_sided(ck, cn - ck, bk, bn - bk)
        print(f"\nRow effect, the clean pair (control − annotation; arms differ ONLY in the row value):")
        print(f"  {row_effect:+.0%}   Fisher exact (two-sided) p = {p_row:.3f}")


if __name__ == "__main__":
    main(sys.argv[1] if len(sys.argv) > 1 else str(pathlib.Path(__file__).parent / "runs"))
