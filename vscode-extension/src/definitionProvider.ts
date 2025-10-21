import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';

export class TypeMuxDefinitionProvider implements vscode.DefinitionProvider {
    provideDefinition(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): vscode.ProviderResult<vscode.Definition | vscode.LocationLink[]> {
        const line = document.lineAt(position.line);
        const lineText = line.text;

        // Check if we're on an import line
        const importMatch = lineText.match(/import\s+"([^"]+)"/);
        if (importMatch) {
            const importPath = importMatch[1];

            // Check if cursor is within the import path string
            const startIndex = lineText.indexOf('"') + 1;
            const endIndex = lineText.lastIndexOf('"');
            if (position.character < startIndex || position.character > endIndex) {
                return null;
            }

            // Resolve the import path relative to the current file
            const currentFileDir = path.dirname(document.uri.fsPath);
            const resolvedPath = path.resolve(currentFileDir, importPath);

            // Check if the file exists
            if (!fs.existsSync(resolvedPath)) {
                return null;
            }

            // Return the location of the imported file
            return new vscode.Location(
                vscode.Uri.file(resolvedPath),
                new vscode.Position(0, 0)
            );
        }

        // Check if we're on a type reference
        const wordRange = document.getWordRangeAtPosition(position);
        if (!wordRange) {
            return null;
        }

        const word = document.getText(wordRange);

        // Check if this looks like a custom type (starts with capital letter)
        if (!/^[A-Z][a-zA-Z0-9]*$/.test(word)) {
            return null;
        }

        // Search for type definition in current file
        const currentFileLocation = this.findTypeDefinition(document, word);
        if (currentFileLocation) {
            return currentFileLocation;
        }

        // Search in imported files
        return this.searchImportedFiles(document, word);
    }

    private findTypeDefinition(document: vscode.TextDocument, typeName: string): vscode.Location | null {
        const text = document.getText();
        const lines = text.split('\n');

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            // Match: type TypeName {, enum TypeName {, or service TypeName {
            const match = line.match(new RegExp(`^\\s*(type|enum|service)\\s+${typeName}\\s*\\{`));
            if (match) {
                return new vscode.Location(
                    document.uri,
                    new vscode.Position(i, match[0].indexOf(typeName))
                );
            }
        }

        return null;
    }

    private searchImportedFiles(document: vscode.TextDocument, typeName: string): vscode.Location | null {
        const text = document.getText();
        const importMatches = text.matchAll(/import\s+"([^"]+)"/g);

        const currentFileDir = path.dirname(document.uri.fsPath);

        for (const match of importMatches) {
            const importPath = match[1];
            const resolvedPath = path.resolve(currentFileDir, importPath);

            if (!fs.existsSync(resolvedPath)) {
                continue;
            }

            try {
                const importedContent = fs.readFileSync(resolvedPath, 'utf-8');
                const lines = importedContent.split('\n');

                for (let i = 0; i < lines.length; i++) {
                    const line = lines[i];
                    const typeMatch = line.match(new RegExp(`^\\s*(type|enum|service)\\s+${typeName}\\s*\\{`));
                    if (typeMatch) {
                        return new vscode.Location(
                            vscode.Uri.file(resolvedPath),
                            new vscode.Position(i, typeMatch[0].indexOf(typeName))
                        );
                    }
                }
            } catch (err) {
                // Skip files we can't read
                continue;
            }
        }

        return null;
    }
}
