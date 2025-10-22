# TypeMUX VS Code Extension - Installation Guide

## Extension Features

✅ **Syntax Highlighting** - Full highlighting for TypeMUX schema files
✅ **Code Snippets** - 20+ snippets for quick code insertion
✅ **Auto-Completion** - Bracket and quote auto-closing
✅ **Code Folding** - Fold type and service definitions
✅ **Indentation** - Automatic indentation

## Installation Methods

### Method 1: Direct Installation (Recommended)

The extension has been installed to your VS Code extensions directory:

```bash
~/.vscode/extensions/typemux-schema-0.1.0/
```

**To activate:**

1. **Restart VS Code** or reload the window:
   - Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on Mac)
   - Type "Developer: Reload Window"
   - Press Enter

2. **Open a `.typemux` file** to test:
   ```bash
   code /home/user/codegen/test-syntax.typemux
   ```

3. **Verify installation:**
   - The file should have syntax highlighting
   - Try typing `type` and pressing Tab to use the snippet

### Method 2: Manual Copy (Alternative)

If the extension isn't working, manually copy it:

```bash
# Create extensions directory if it doesn't exist
mkdir -p ~/.vscode/extensions/

# Copy the extension
cp -r /home/user/codegen/vscode-extension ~/.vscode/extensions/typemux-schema-0.1.0

# Restart VS Code
```

### Method 3: Using VS Code CLI (If available)

If you have the VS Code CLI installed:

```bash
cd /home/user/codegen/vscode-extension
code --install-extension $(pwd)
```

## Testing the Extension

### 1. Open Test File

```bash
code /home/user/codegen/test-syntax.typemux
```

### 2. Test Syntax Highlighting

You should see:
- **Keywords** (enum, type, service, rpc) highlighted
- **Types** (string, int32, timestamp) highlighted
- **Attributes** (@required, @exclude) highlighted
- **Comments** (///, //) styled differently
- **Numbers** highlighted

### 3. Test Snippets

Try these snippets (type and press Tab):

| Snippet | Description |
|---------|-------------|
| `type` | Create a type definition |
| `enum` | Create an enum |
| `service` | Create a service |
| `fieldnum` | Create a field with custom number |
| `rpc` | Create an RPC method |
| `doc` | Add documentation comment |
| `http` | Add HTTP annotation |
| `schema` | Complete schema template |

**Example:**
1. Create a new `.typemux` file
2. Type `type` and press `Tab`
3. Fill in the placeholders

### 4. Test Auto-Completion

- Type `{` - should auto-close with `}`
- Type `(` - should auto-close with `)`
- Type `"` - should auto-close with `"`

## Troubleshooting

### Extension Not Showing

1. **Check installation:**
   ```bash
   ls -la ~/.vscode/extensions/typemux-schema-0.1.0/
   ```

2. **Verify files exist:**
   - `package.json`
   - `language-configuration.json`
   - `syntaxes/typemux.tmLanguage.json`
   - `snippets/typemux.json`

3. **Check VS Code version:**
   - Requires VS Code 1.70.0 or higher
   - Check with: `code --version`

### Syntax Highlighting Not Working

1. **Verify file extension:**
   - File must end with `.typemux`
   - Check with: `ls -l *.typemux`

2. **Force language mode:**
   - Click language in bottom-right of VS Code
   - Select "TypeMUX Schema"

3. **Check for conflicts:**
   - Other extensions might interfere
   - Try disabling other language extensions

### Snippets Not Working

1. **Enable snippets in settings:**
   - Press `Ctrl+,` (or `Cmd+,`)
   - Search for "snippets"
   - Ensure "Editor: Snippet Suggestions" is enabled

2. **Try trigger:**
   - Type the snippet prefix
   - Press `Tab` (not Enter)
   - Make sure cursor is at the start of a line for some snippets

## Uninstallation

To remove the extension:

```bash
rm -rf ~/.vscode/extensions/typemux-schema-0.1.0
```

Then reload VS Code.

## Development

To modify the extension:

1. **Edit files:**
   ```bash
   cd /home/user/codegen/vscode-extension
   ```

2. **Update syntax:**
   - Edit `syntaxes/typemux.tmLanguage.json`

3. **Update snippets:**
   - Edit `snippets/typemux.json`

4. **Reinstall:**
   ```bash
   cp -r /home/user/codegen/vscode-extension ~/.vscode/extensions/typemux-schema-0.1.0
   ```

5. **Reload VS Code:**
   - `Ctrl+Shift+P` → "Developer: Reload Window"

## Quick Reference

### File Structure
```
~/.vscode/extensions/typemux-schema-0.1.0/
├── package.json                    # Extension metadata
├── language-configuration.json     # Brackets, comments, etc.
├── README.md                       # Documentation
├── syntaxes/
│   └── typemux.tmLanguage.json    # Syntax highlighting rules
└── snippets/
    └── typemux.json               # Code snippets
```

### Example Schema Files

Test files are available in:
- `/home/user/codegen/examples/basic.typemux`
- `/home/user/codegen/examples/custom_field_numbers.typemux`
- `/home/user/codegen/examples/example.typemux`
- `/home/user/codegen/test-syntax.typemux`

## Support

For issues or feature requests, visit:
https://github.com/rasmartins/typemux
