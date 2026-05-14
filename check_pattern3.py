#!/usr/bin/env python3
import re

# Try building the pattern programmatically
p = r'name:\s*"([^"]+)"\s*,\s*json:\s*\[\]byte\('
backtick = '`'
p += backtick
p += r'([^`]*)'
p += backtick
p += r'\)\s*,\s*wantErr:\s*(true|false)'
print(f"Pattern: {p}")
try:
    compiled = re.compile(p)
    print("OK - compiles")
except Exception as e:
    print(f"Error: {e}")