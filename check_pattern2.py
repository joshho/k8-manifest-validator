backtick_pattern = r'name:\s*"([^"]+)"\s*,\s*json:\s*\[\]byte\()`([^`]*)`\)\s*,\s*wantErr:\s*(true|false)'
print(backtick_pattern)
print(f"Position 42: '{backtick_pattern[40:45]}")