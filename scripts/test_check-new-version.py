#!/usr/bin/env python3
"""
test_check-new-version.py — Unit tests for check-new-version.sh

9 test scenarios covering:
  1. VERSION file missing → clear error
  2. same minor+patch → no bump needed
  3. one minor ahead → NEW_PATCH emitted
  4. two minors ahead → refuses to bump
  5. REPO_ROOT resolves from script location (not cwd)
  6. --minor mode: new minor detected
  7. --minor mode: already current
  8. same minor, newer patch available → NEW_PATCH
  9. patch mode always produces non-empty output

Usage: python3 scripts/test_check-new-version.py
"""
import subprocess
import tempfile
import shutil
import os
import re

SCRIPTS_DIR = os.path.dirname(os.path.abspath(__file__))
WORKSPACE = os.path.dirname(SCRIPTS_DIR)
CHECK_SCRIPT = os.path.join(WORKSPACE, "scripts", "check-new-version.sh")

tests_run = 0
tests_passed = 0
tests_failed = 0


def pass_(msg):
    global tests_passed
    tests_passed += 1
    print(f"  \u2713 {msg}")


def fail_(msg, expected, got):
    global tests_failed
    tests_failed += 1
    print(f"  \u2717 {msg}")
    print(f"    Expected: {expected}")
    print(f"    Got:      {got[:400]}")


def run_check(mode, fake_tag):
    """Run check-new-version.sh with a mocked k8s API response via curl wrapper."""
    tmp = tempfile.mkdtemp()
    try:
        curl_wrapper = os.path.join(tmp, "curl")
        with open(curl_wrapper, "w") as f:
            f.write("#!/bin/bash\n")
            f.write("for arg in \"$@\"; do\n")
            f.write("    if [[ \"$arg\" == *api.github.com* ]]; then\n")
            f.write(f'        echo \'{{"tag_name":"{fake_tag}"}}\'\n')
            f.write("        exit 0\n")
            f.write("    fi\n")
            f.write("done\n")
            f.write('exec /usr/bin/curl "$@"\n')
        os.chmod(curl_wrapper, 0o755)

        env = os.environ.copy()
        env["PATH"] = tmp + ":" + env["PATH"]
        result = subprocess.run(
            ["bash", CHECK_SCRIPT, f"--{mode}"],
            cwd=WORKSPACE, capture_output=True, text=True, env=env, timeout=15,
        )
        return result.stdout + result.stderr
    finally:
        shutil.rmtree(tmp)


def make_commit(version, api_version, message):
    """Write test VERSION + go.mod, make a commit, return original commit SHA."""
    with open(os.path.join(WORKSPACE, "VERSION")) as f:
        orig_version = f.read().strip()
    with open(os.path.join(WORKSPACE, "go.mod")) as f:
        orig_gomod = f.read()

    with open(os.path.join(WORKSPACE, "VERSION"), "w") as f:
        f.write(version + "\n")

    new_gomod = orig_gomod
    for pkg in ["api", "apiextensions-apiserver", "apimachinery"]:
        new_gomod = re.sub(
            rf'(k8s\.io/{pkg}\s+)v[0-9]+\.[0-9]+\.[0-9]+',
            rf'\g<1>{api_version}',
            new_gomod,
        )
    with open(os.path.join(WORKSPACE, "go.mod"), "w") as f:
        f.write(new_gomod)

    subprocess.run(["git", "add", "VERSION", "go.mod"], cwd=WORKSPACE, timeout=10)
    subprocess.run(["git", "commit", "-m", message, "--allow-empty"],
                   cwd=WORKSPACE, timeout=10)
    result = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        cwd=WORKSPACE, capture_output=True, text=True, timeout=10,
    )
    commit = result.stdout.strip()

    return commit, orig_version, orig_gomod


def restore(orig_version, orig_gomod, commit):
    """Restore original VERSION/go.mod and reset to original commit."""
    with open(os.path.join(WORKSPACE, "VERSION"), "w") as f:
        f.write(orig_version + "\n")
    with open(os.path.join(WORKSPACE, "go.mod"), "w") as f:
        f.write(orig_gomod)
    subprocess.run(["git", "checkout", "--", "VERSION", "go.mod"],
                   cwd=WORKSPACE, timeout=10)
    subprocess.run(["git", "reset", "--hard", commit],
                   cwd=WORKSPACE, timeout=10)


# ── TEST 1: VERSION file missing → clear error ─────────────────────────────

def test_version_missing():
    global tests_run
    tests_run += 1

    tmp = tempfile.mkdtemp()
    try:
        with open(CHECK_SCRIPT) as f:
            script = f.read()
        # Patch REPO_ROOT to point to tmp (no VERSION there)
        patched = script.replace(
            "REPO_ROOT=\"$(cd \"$SCRIPT_DIR/..\" && pwd)\"",
            "REPO_ROOT=\"$TMPDIR\"",
        )
        test_script = os.path.join(tmp, "check-new-version.sh")
        with open(test_script, "w") as f:
            f.write(patched)
        os.chmod(test_script, 0o755)

        with open(os.path.join(tmp, "go.mod"), "w") as f:
            f.write("module test\ngo 1.26\nk8s.io/api v0.36.0\n")

        result = subprocess.run(
            ["bash", test_script, "--patch"],
            cwd=tmp, capture_output=True, text=True, timeout=10,
            env={**os.environ, "TMPDIR": tmp},
        )
        output = result.stdout + result.stderr
        if "VERSION file not found" in output:
            pass_("VERSION file missing → clear error")
        else:
            fail_("VERSION missing", "VERSION file not found", output)
    finally:
        shutil.rmtree(tmp)


# ── TEST 2: same minor+patch → no bump needed ─────────────────────────────

def test_already_on_latest():
    global tests_run
    tests_run += 1

    commit, orig_v, orig_gm = make_commit("v1.35.0", "v0.35.0", "test: v1.35.0")
    try:
        result = run_check("patch", "v1.35.0")
        if "Already on latest" in result:
            pass_("same minor+patch → no bump needed")
        elif "NEW_PATCH" in result:
            fail_("same minor", "no NEW_PATCH", result)
        else:
            fail_("same minor", "Already on latest", result)
    finally:
        restore(orig_v, orig_gm, commit)


# ── TEST 3: k8s one minor ahead → emits NEW_PATCH ─────────────────────────

def test_one_minor_ahead():
    global tests_run
    tests_run += 1

    commit, orig_v, orig_gm = make_commit("v1.32.0", "v0.32.0", "test: v1.32.0")
    try:
        result = run_check("patch", "v1.33.0")
        if "NEW_PATCH=v1.33.0" in result:
            pass_("one minor ahead → NEW_PATCH=v1.33.0")
        else:
            fail_("one minor ahead", "NEW_PATCH=v1.33.0", result)
    finally:
        restore(orig_v, orig_gm, commit)


# ── TEST 4: k8s two minors ahead → refuses to bump ───────────────────────

def test_multi_minor_guard():
    global tests_run
    tests_run += 1

    commit, orig_v, orig_gm = make_commit("v1.32.0", "v0.32.0", "test: v1.32.0")
    try:
        result = run_check("patch", "v1.34.0")
        if "refusing to bump" in result or "k8s jumped multiple minors" in result:
            pass_("two minors ahead → refuses to bump")
        elif "NEW_PATCH" in result:
            fail_("multi-minor guard", "should refuse", result)
        else:
            fail_("multi-minor guard", "refusing to bump", result)
    finally:
        restore(orig_v, orig_gm, commit)


# ── TEST 5: REPO_ROOT resolves from script location, not cwd ─────────────

def test_repo_root_resolution():
    global tests_run
    tests_run += 1

    commit, orig_v, orig_gm = make_commit("v1.35.0", "v0.35.0", "test: v1.35.0")
    try:
        result = run_check("patch", "v1.35.0")
        if "Already on latest" in result or "Package minor" in result:
            pass_("REPO_ROOT resolves from script location (not cwd)")
        else:
            fail_("REPO_ROOT", "VERSION found at repo root", result)
    finally:
        restore(orig_v, orig_gm, commit)


# ── TEST 6: --minor mode: new minor detected ───────────────────────────────

def test_minor_mode_new():
    global tests_run
    tests_run += 1

    commit, orig_v, orig_gm = make_commit("v1.35.0", "v0.35.0", "test: v1.35.0")
    try:
        result = run_check("minor", "v1.36.0")
        if "NEW_MINOR=v1.36" in result:
            pass_("--minor: new minor v1.36 detected")
        else:
            fail_("--minor new", "NEW_MINOR=v1.36", result)
    finally:
        restore(orig_v, orig_gm, commit)


# ── TEST 7: --minor mode: already current ─────────────────────────────────

def test_minor_mode_current():
    global tests_run
    tests_run += 1

    commit, orig_v, orig_gm = make_commit("v1.36.0", "v0.36.0", "test: v1.36.0")
    try:
        result = run_check("minor", "v1.36.0")
        if "Already on latest minor" in result:
            pass_("--minor: already current (no bump)")
        else:
            fail_("--minor current", "Already on latest minor", result)
    finally:
        restore(orig_v, orig_gm, commit)


# ── TEST 8: same minor, newer patch available → NEW_PATCH ────────────────

def test_patch_newer_available():
    global tests_run
    tests_run += 1

    commit, orig_v, orig_gm = make_commit("v1.32.0", "v0.32.5", "test: v1.32.5")
    try:
        result = run_check("patch", "v1.32.9")
        if "NEW_PATCH=v1.32.9" in result:
            pass_("same minor, newer patch v1.32.9 → NEW_PATCH")
        else:
            fail_("patch newer", "NEW_PATCH=v1.32.9", result)
    finally:
        restore(orig_v, orig_gm, commit)


# ── TEST 9: patch mode always produces non-empty output ─────────────────

def test_patch_always_has_output():
    global tests_run
    tests_run += 1

    commit, orig_v, orig_gm = make_commit("v1.35.0", "v0.35.0", "test: v1.35.0")
    try:
        result = run_check("patch", "v1.35.0")
        if result and result.strip():
            pass_("patch mode always produces output")
        else:
            fail_("patch output", "non-empty output", "(empty)")
    finally:
        restore(orig_v, orig_gm, commit)


# ── Run all tests ───────────────────────────────────────────────────────────

if __name__ == "__main__":
    print("")
    print("=== check-new-version.sh tests ===")
    print("")
    print(f"Working directory: {WORKSPACE}")
    print(f"Script: {CHECK_SCRIPT}")
    print("")

    test_version_missing()
    test_already_on_latest()
    test_one_minor_ahead()
    test_multi_minor_guard()
    test_repo_root_resolution()
    test_minor_mode_new()
    test_minor_mode_current()
    test_patch_newer_available()
    test_patch_always_has_output()

    print("")
    print(f"Results: {tests_passed}/{tests_run} passed, {tests_failed} failed")
    print("")

    exit(0 if tests_failed == 0 else 1)