#!/usr/bin/env python3
import re
content = open('scripts/extract_fixtures.py').read()
# Check if backtick pattern is correct
backtick_pattern = r'(name:\s*"[^"]+"\s*,\s*json:\s*\[\]byte\()`([^`]*)`\),\s*wantErr:\s*(true|false)'
print('Pattern check...')
try:
    re.compile(backtick_pattern)
    print('Pattern compiles OK')
except Exception as e:
    print(f'Pattern error: {e}')